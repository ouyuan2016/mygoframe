package queue

import (
	"github.com/hibiken/asynq"
)

// NewClient creates and returns a new asynq client.
func NewClient(opt asynq.RedisClientOpt) *asynq.Client {
	client := asynq.NewClient(opt)
	return client
}

// NewServer creates and returns a new asynq server.
func NewServer(opt asynq.RedisClientOpt, concurrency int, queues map[string]int) *asynq.Server {
	srv := asynq.NewServer(
		opt,
		asynq.Config{
			Concurrency: concurrency,
			Queues:      queues,
		},
	)
	return srv
}

// NewServeMux creates and returns a new asynq ServeMux.
func NewServeMux() *asynq.ServeMux {
	return asynq.NewServeMux()
}
