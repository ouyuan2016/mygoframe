package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TypeWelcomeEmail   = "email:welcome"
	TypeHelloWorld     = "task:hello_world"
	TypeSendLaterEmail = "email:send_later"
)

// NewWelcomeEmailTask creates a new welcome email task.
func NewWelcomeEmailTask(userID int) (*asynq.Task, error) {
	payload, err := json.Marshal(map[string]interface{}{"user_id": userID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeWelcomeEmail, payload), nil
}

// HandleWelcomeEmailTask handles the welcome email task.
func HandleWelcomeEmailTask(ctx context.Context, t *asynq.Task) error {
	var p map[string]interface{}
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	userID := p["user_id"]
	log.Printf("Sending a welcome email to user %v", userID)
	return nil
}

// NewHelloWorldTask creates a new hello world task.
func NewHelloWorldTask() (*asynq.Task, error) {
	return asynq.NewTask(TypeHelloWorld, nil), nil
}

// HandleHelloWorldTask handles the hello world task.
func HandleHelloWorldTask(ctx context.Context, t *asynq.Task) error {
	log.Println("hello world")
	return nil
}

// SendLaterEmailPayload defines the payload for the send later email task.
type SendLaterEmailPayload struct {
	UserID string    `json:"user_id"`
	SendAt time.Time `json:"send_at"`
}

// NewSendLaterEmailTask creates a new task to send an email later.
func NewSendLaterEmailTask(userID string, sendAt time.Time) (*asynq.Task, error) {
	payload, err := json.Marshal(SendLaterEmailPayload{UserID: userID, SendAt: sendAt})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendLaterEmail, payload), nil
}

// HandleSendLaterEmailTask handles the send later email task.
func HandleSendLaterEmailTask(ctx context.Context, t *asynq.Task) error {
	var p SendLaterEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	log.Printf("Sending Delayed Email to User ID %s, which was scheduled at %s", p.UserID, p.SendAt)
	// Here you would put your email sending logic.
	return nil
}
