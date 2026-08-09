package domain

import "context"

type Notification struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority int    `json:"priority"` // e.g. 1 (low) to 10 (high/urgent)
}

type NotificationService interface {
	Send(ctx context.Context, title, message string, priority int) error
}
