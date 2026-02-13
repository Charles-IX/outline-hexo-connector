package outline

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"outline-hexo-connector/internal/config"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	cfg                  *config.Config
	httpClient           *http.Client
	httpClientNoRedirect *http.Client
	justCreatedDocs      sync.Map
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		httpClientNoRedirect: &http.Client{
			Timeout: time.Second * 10,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (c *Client) verifyWebhook(body []byte, signatureHeader string) error {
	if c.cfg.OutlineWebhookSecret == "" {
		return fmt.Errorf("No webhook secret configured in config")
	}

	var t, s string
	pairs := strings.Split(signatureHeader, ",")
	for _, pair := range pairs {
		key, value, found := strings.Cut(pair, "=")
		if !found {
			return fmt.Errorf("Invalid signature header format")
		}
		switch key {
		case "t":
			t = value
		case "s":
			s = value
		}
	}
	if t == "" || s == "" {
		return fmt.Errorf("Missing fields in signature header")
	}

	ms, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid timestamp format in signature header - %w", err)
	}
	timestamp := time.UnixMilli(ms)
	diff := time.Since(timestamp)
	if diff > time.Minute*2 {
		return fmt.Errorf("Signature timestamp is too old")
	} else if diff < -time.Second*30 {
		return fmt.Errorf("Signature timestamp is in the future")
	}

	message := t + "." + string(body)
	mac := hmac.New(sha256.New, []byte(c.cfg.OutlineWebhookSecret))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(s)) {
		return fmt.Errorf("Signature invalid")
	}

	return nil
}

func (c *Client) parseWebhook(body []byte) (*Webhook, error) {
	var webhook Webhook
	err := json.Unmarshal(body, &webhook)
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

func (c *Client) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	log.Println("Received webhook request:")

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 10*1024*1024))
	if err != nil {
		log.Printf("Error reading webhook - %v", err)
		http.Error(w, "Error reading webhook", http.StatusBadRequest)
		return
	}

	err = c.verifyWebhook(body, r.Header.Get("Outline-Signature"))
	if err != nil {
		log.Printf("Error verifying webhook - %v", err)
		http.Error(w, "Error verifying webhook", http.StatusUnauthorized)
		return
	}

	webhook, err := c.parseWebhook(body)
	if err != nil {
		log.Printf("Error parsing webhook - %v", err)
		http.Error(w, "Error parsing webhook", http.StatusBadRequest)
		return
	}

	fmt.Printf("Event: %s\nDocument ID: %s\nDocument Title: %s\n", webhook.Event, webhook.Payload.Model.ID, webhook.Payload.Model.Title)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Acknowledged"))

	if webhook.Payload.Model.ParentDocumentID != "" {
		parentDocument, err := c.GetDocument(webhook.Payload.Model.ParentDocumentID)
		if err != nil {
			log.Printf("Error fetching parent document info - %v", err)
			return
		}
		webhook.Payload.Model.ParentDocument = &parentDocument

	} else {
		webhook.Payload.Model.ParentDocument = &DocumentPayload{
			Title: "null",
		}
	}
	fmt.Printf("Parent Name: %s\n", webhook.Payload.Model.ParentDocument.Title)

	collection, err := c.GetCollection(webhook.Payload.Model.CollectionID)
	if err != nil {
		log.Printf("Error fetching collection info - %v", err)
		return
	}
	fmt.Printf("Collection Name: %s\n", collection.Name)

	if collection.Name != c.cfg.OutlineCollectionUsedForBlog {
		log.Printf("Not desired collection - Skipping")
		return
	}

	switch webhook.Event {
	case "documents.create":
		log.Printf("This is a spaceholder to prevent my IDE from yelling.")
		// TODO:
		// Outline sends this whenever a document is created, and is always
		// followed by a documents.publish event. The following event
		// must be ignored, for we don't want to trigger a Hexo build when
		// the document is empty.

	case "documents.publish":
		fallthrough
	case "documents.unarchive":
		fallthrough
	case "documents.restore":
		log.Printf("This is a spaceholder to prevent my IDE from yelling.")
		Blog := &Document{
			ID:        webhook.Payload.Model.ID,
			Title:     webhook.Payload.Model.Title,
			Text:      webhook.Payload.Model.Text,
			CreatedAt: webhook.Payload.Model.CreatedAt,
			UpdatedAt: webhook.Payload.Model.UpdatedAt,
			Category:  webhook.Payload.Model.ParentDocument.Title,
		}
		log.Printf("Document published - ID: %s, Title: %s, Category: %s", Blog.ID, Blog.Title, Blog.Category)
		// TODO:
		// Add the corresponding .md then trigger Hexo build. Or better,
		// put it in a queue for a periodical Hexo build to consume.

	case "documents.unpublish":
		fallthrough
	case "documents.archive":
		fallthrough
	case "documents.delete":
		log.Printf("This is a spaceholder to prevent my IDE from yelling.")
		// TODO:
		// Remove the corresponding .md then trigger Hexo build or set a flag
		// and wait for the periodical Hexo build to consume.

	case "documents.update":
		log.Printf("This is a spaceholder to prevent my IDE from yelling.")
		// TODO:
		// Am I really going to do shit about this?
		// If really, who is going to validate whether the document is ready
		// for publishing to Hexo?
		// Or I'm just too drowsy.

	// TODO:
	// Better do something to handle documents.move and documents.title_change.
	// I don't see any difference between these two and documents.publish,
	// for we want to use Outline's document ID for Hexo's .md filename.
	// Every "update" is indeed an overwrite handled by os.

	default:
		log.Printf("Unhandled event type: %s", webhook.Event)
	}
}

func (c *Client) newRequest(endpoint string, reqPayload any) (*http.Request, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("Error encoding request payload - %w", err)
	}

	req, err := http.NewRequest("POST", c.cfg.OutlineAPIURL+endpoint, &buf)
	if err != nil {
		return nil, fmt.Errorf("Error creating request - %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.OutlineAPIKey)
	return req, nil
}

func getInfoByID[T any](c *Client, endpoint string, id string) (T, error) {
	reqPayload := RequestPayload{ID: id}
	var zero T

	req, err := c.newRequest(endpoint, reqPayload)
	if err != nil {
		return zero, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return zero, fmt.Errorf("Error requesting Outline API - %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&apiErr)
		if err != nil {
			return zero, fmt.Errorf("Error decoding API error response - %w", err)
		}
		return zero, fmt.Errorf("Unexpected API http status - %d, %s", resp.StatusCode, apiErr.Error())
	}

	var response struct {
		Data T `json:"data"`
	}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&response)
	if err != nil {
		return zero, fmt.Errorf("Error decoding API response - %w", err)
	}
	return response.Data, nil
}

func (c *Client) GetDocument(id string) (DocumentPayload, error) {
	return getInfoByID[DocumentPayload](c, "/documents.info", id)
}

func (c *Client) GetCollection(id string) (CollectionPayload, error) {
	return getInfoByID[CollectionPayload](c, "/collections.info", id)
}

func (c *Client) GetAttachmentUrl(id string) (string, error) {
	reqPayload := RequestPayload{ID: id}
	req, err := c.newRequest("/attachments.redirect", reqPayload)
	if err != nil {
		return "", fmt.Errorf("Error creating request - %w", err)
	}

	resp, err := c.httpClientNoRedirect.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error requesting Outline API - %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		var apiErr APIError
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&apiErr)
		if err != nil {
			return "", fmt.Errorf("Error decoding API error response - %w", err)
		}
		return "", fmt.Errorf("Unexpected API http status - %d, %s", resp.StatusCode, apiErr.Error())
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("Missing Location header in API response")
	}
	return location, nil
}

func (c *Client) unpublishDocument(id string) error {
	reqPayload := RequestPayload{ID: id}
	req, err := c.newRequest("/documents.unpublish", reqPayload)
	if err != nil {
		return fmt.Errorf("Error creating request - %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Error requesting Outline API - %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&apiErr)
		if err != nil {
			return fmt.Errorf("Error decoding API error response - %w", err)
		}
		return fmt.Errorf("Unexpected API http status - %d, %s", resp.StatusCode, apiErr.Error())
	}
	return nil
}
