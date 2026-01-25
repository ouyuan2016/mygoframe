package task

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"mygoframe/pkg/queue"

	"github.com/hibiken/asynq"
)

func NewWelcomeEmailTask(userID int) (*asynq.Task, error) {
	payload, err := json.Marshal(map[string]interface{}{"user_id": userID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeWelcomeEmail, payload), nil
}

func HandleWelcomeEmailTask(ctx context.Context, t *asynq.Task) error {
	var p map[string]interface{}
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	userID := p["user_id"]
	log.Printf("Sending a welcome email to user %v", userID)
	return nil
}

func EnqueueWelcomeEmailTask(userID int, queueName string) (*asynq.TaskInfo, error) {
	task, err := NewWelcomeEmailTask(userID)
	if err != nil {
		return nil, fmt.Errorf("创建欢迎邮件任务失败: %w", err)
	}

	info, err := queue.Client.Enqueue(task, asynq.MaxRetry(3), asynq.Timeout(20*time.Minute), asynq.Queue(queueName))
	if err != nil {
		return nil, fmt.Errorf("任务入队失败: %w", err)
	}

	return info, nil
}

type SendLaterEmailPayload struct {
	UserID string    `json:"user_id"`
	SendAt time.Time `json:"send_at"`
}

func NewSendLaterEmailTask(userID string, sendAt time.Time) (*asynq.Task, error) {
	payload, err := json.Marshal(SendLaterEmailPayload{UserID: userID, SendAt: sendAt})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendLaterEmail, payload), nil
}

func HandleSendLaterEmailTask(ctx context.Context, t *asynq.Task) error {
	var p SendLaterEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	log.Printf("Sending Delayed Email to User ID %s, which was scheduled at %s", p.UserID, p.SendAt)
	return nil
}

func EnqueueSendLaterEmailTask(userID string, delay time.Duration) (*asynq.TaskInfo, error) {
	task, err := NewSendLaterEmailTask(userID, time.Now().Add(delay))
	if err != nil {
		return nil, fmt.Errorf("创建延迟邮件任务失败: %w", err)
	}

	info, err := queue.Client.Enqueue(task, asynq.ProcessIn(delay))
	if err != nil {
		return nil, fmt.Errorf("延迟任务入队失败: %w", err)
	}

	return info, nil
}
