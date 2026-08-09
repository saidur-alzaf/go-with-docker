package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go-sqlite-api/internal/domain"
)

type gotifyService struct {
	url        string
	token      string
	httpClient *http.Client
}

func NewGotifyService(url, token string) domain.NotificationService {
	return &gotifyService{
		url:   url,
		token: token,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type gotifyPayload struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`
}

func (g *gotifyService) Send(ctx context.Context, title, message string, priority int) error {
	if g.url == "" || g.token == "" {
		log.Printf("[Gotify Skipped] URL or Token not configured. Message: '%s - %s'", title, message)
		return nil
	}

	payload := gotifyPayload{
		Title:    title,
		Message:  message,
		Priority: priority,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal gotify payload: %w", err)
	}

	endpoint := fmt.Sprintf("%s/message?token=%s", g.url, g.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create gotify request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		log.Printf("[Gotify Error] Failed to send notification '%s': %v", title, err)
		return fmt.Errorf("gotify http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("[Gotify Error] Gotify responded with status code %d for '%s'", resp.StatusCode, title)
		return fmt.Errorf("gotify server returned error code %d", resp.StatusCode)
	}

	log.Printf("[Gotify Sent] Title: '%s', Priority: %d", title, priority)
	return nil
}
