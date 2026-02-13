package outline

import (
	"fmt"
)

type Webhook struct {
	WebhookSubscriptionID string `json:"webhookSubscriptionId"`
	Event                 string `json:"event"`
	Payload               struct {
		Model DocumentPayload `json:"model"`
	} `json:"payload"`
}

type DocumentPayload struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Text             string `json:"text"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
	CollectionID     string `json:"collectionId"`
	ParentDocumentID string `json:"parentDocumentId"`
	ParentDocument   *DocumentPayload
	Collection       *CollectionPayload
}

type CollectionPayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Document struct {
	ID        string
	Title     string
	Text      string
	CreatedAt string
	UpdatedAt string
	Category  string
	Tags      string
}

type RequestPayload struct {
	ID string `json:"id"`
}

type APIError struct {
	Err     string `json:"error"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s - %s", e.Err, e.Message)
	}
	return e.Err
}
