package queue

import (
	"context"
	"fmt"
	"mygoframe/pkg/config"

	"github.com/hibiken/asynq"
)

var (
	Client     *asynq.Client
	Server     *asynq.Server
	Mux        *asynq.ServeMux
	Scheduler  *asynq.Scheduler
	HandleFunc = make(map[string]func(context.Context, *asynq.Task) error)
)

func getRedisOpt() asynq.RedisClientOpt {
	redisConf := config.GetConfig().Redis
	return asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", redisConf.Host, redisConf.Port),
		Password: redisConf.Password,
		DB:       redisConf.DB,
	}
}

// InitQueue initializes the asynq client.
func InitQueue() {
	Client = NewClient(getRedisOpt())
}

// InitQueueServer initializes the asynq server.
func InitQueueServer() {
	cfg := config.GetConfig().Queue
	Server = NewServer(getRedisOpt(), cfg.Concurrency, cfg.Queues)
	Mux = NewServeMux()
	Scheduler = asynq.NewScheduler(getRedisOpt(), nil)
}

// RegisterHandler registers a handler for a given task type.
func RegisterHandler(pattern string, handler func(context.Context, *asynq.Task) error) {
	HandleFunc[pattern] = handler
}

// RegisterCronJob registers a cron job.
func RegisterCronJob(spec string, task *asynq.Task) (string, error) {
	return Scheduler.Register(spec, task)
}

// StopScheduler stops the scheduler.
func StopScheduler() {
	if Scheduler != nil {
		Scheduler.Shutdown()
	}
}
