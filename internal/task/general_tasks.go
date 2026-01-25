package task

import (
	"context"
	"log"

	"github.com/hibiken/asynq"
)

func NewHelloWorldTask() *asynq.Task {
	return asynq.NewTask(CronHelloWorld, nil)
}

func HandleHelloWorldTask(ctx context.Context, t *asynq.Task) error {
	log.Println("hello world")
	return nil
}
