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
	"outline-hexo-connector/internal/hexo"
	"outline-hexo-connector/internal/processor"
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
		return err
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

func (c *Client) logWebhook(webhook *Webhook) {
	log.Println("Received webhook request:")
	fmt.Printf("Event: %s\nDocument ID: %s\nDocument Title: %s\n", webhook.Event, webhook.Payload.Model.ID, webhook.Payload.Model.Title)
	fmt.Printf("Parent Name: %s\n", webhook.Payload.Model.ParentDocument.Title)
	fmt.Printf("Collection Name: %s\n", webhook.Payload.Model.Collection.Name)
}

func (c *Client) HandleWebhook(w http.ResponseWriter, r *http.Request) {
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

	collection, err := c.GetCollection(webhook.Payload.Model.CollectionID)
	webhook.Payload.Model.Collection = &collection
	if err != nil {
		log.Printf("Error fetching collection info - %v", err)
		return
	}

	if collection.Name != c.cfg.OutlineCollectionUsedForBlog {
		// log.Printf("Not desired collection - Skipping")
		// Commented out to reduce log noise
		return
	}

	switch webhook.Event {
	case "documents.create":
		c.logWebhook(webhook)
		c.justCreatedDocs.Store(webhook.Payload.Model.ID, true)
		go c.unpublishDocument(webhook.Payload.Model.ID)
		time.AfterFunc(time.Second*10, func() {
			c.justCreatedDocs.Delete(webhook.Payload.Model.ID)
		})

	case "documents.publish":
		_, justCreated := c.justCreatedDocs.Load(webhook.Payload.Model.ID)
		if justCreated {
			log.Printf("Document just created - Ignoring publish event")
			return
		}
		fallthrough
	case "documents.unarchive":
		fallthrough
	case "documents.restore":
		c.logWebhook(webhook)

		post := &hexo.Post{
			ID:       webhook.Payload.Model.ID,
			Title:    webhook.Payload.Model.Title,
			Date:     webhook.Payload.Model.CreatedAt,
			Updated:  webhook.Payload.Model.UpdatedAt,
			Category: webhook.Payload.Model.ParentDocument.Title,
			Content:  webhook.Payload.Model.Text,
		}
		post.Content, err = processor.ConvertAttachmentUrl(c, post.Content)
		if err != nil {
			log.Printf("Error converting attachment URLs - %v", err)
			return
		}
		metadataAndText := processor.ExtractMetadataAndText(post.Content)
		post.BannerImg = metadataAndText.BannerImg
		post.IndexImg = metadataAndText.IndexImg
		post.Tags = metadataAndText.Tags
		post.Content = metadataAndText.Text

		fmt.Printf("Converted blog content:\n%s\n", post.Content)
		// TODO:
		// Add the corresponding .md then trigger Hexo build. Or better,
		// put it in a queue for a periodical Hexo build to consume.

	case "documents.unpublish":
		_, justCreated := c.justCreatedDocs.Load(webhook.Payload.Model.ID)
		if justCreated {
			log.Printf("Document just created - Ignoring unpublish event")
			c.justCreatedDocs.Delete(webhook.Payload.Model.ID)
			return
		}
		fallthrough
	case "documents.archive":
		fallthrough
	case "documents.delete":
		log.Printf("This is a spaceholder to prevent my IDE from yelling.")
		// TODO:
		// Remove the corresponding .md then trigger Hexo build or set a flag
		// and wait for the periodical Hexo build to consume.

	case "documents.update":
		return
		// TODO:
		// It's simple. Just treat it as a create event, and send documents.unpublish
		// to Outline API so that we don't send some draft version to Hexo.
		// Let draft be draft. But are we going to unpublish it in Hexo?
		// If implemented, we can use documents.archive
		// Or documets.move or documents.delete to unpublish in Hexo.
		// Drafts also send documents.update event. But draft's webhook has "publishedAt": null
		// We can use that to identify a draft.

		// Thankfully one cannot update an archived document, so no need to check that.

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
		return nil, err
	}

	req, err := http.NewRequest("POST", c.cfg.OutlineAPIURL+endpoint, &buf)
	if err != nil {
		return nil, err
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
		return zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&apiErr)
		if err != nil {
			return zero, err
		}
		return zero, fmt.Errorf("Unexpected API http status - %d, %s", resp.StatusCode, apiErr.Error())
	}

	var response struct {
		Data T `json:"data"`
	}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&response)
	if err != nil {
		return zero, err
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
		return "", err
	}

	resp, err := c.httpClientNoRedirect.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		var apiErr APIError
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&apiErr)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("Unexpected API http status - %d, %s", resp.StatusCode, apiErr.Error())
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("Missing Location header in API response")
	}

	// Cut off S3 presigned part, the bucket should be public-readable
	if index := strings.Index(location, "?"); index != -1 {
		return location[:index], nil
	}
	return location, nil
}

func (c *Client) unpublishDocument(id string) error {
	reqPayload := RequestPayload{ID: id}
	req, err := c.newRequest("/documents.unpublish", reqPayload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&apiErr)
		if err != nil {
			return err
		}
		return fmt.Errorf("Unexpected API http status - %d, %s", resp.StatusCode, apiErr.Error())
	}
	return nil
}
