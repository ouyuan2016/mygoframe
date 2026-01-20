package queue

import "context"

// Tasker defines the interface for a task that can be processed by the queue.
type Tasker interface {
	Type() string
	Payload() []byte
}

// TaskHandler defines the interface for a task handler.
type TaskHandler interface {
	ProcessTask(context.Context, Tasker) error
}

// Queuer defines the interface for a queue.
type Queuer interface {
	Enqueue(Tasker) error
	RegisterHandler(string, TaskHandler)
	Start()
	Shutdown()
}
