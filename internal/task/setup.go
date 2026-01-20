package task

import (
	"log"
	"mygoframe/pkg/queue"
)

func Setup() {
	queue.RegisterHandler(TypeWelcomeEmail, HandleWelcomeEmailTask)
	queue.RegisterHandler(TypeHelloWorld, HandleHelloWorldTask)
	queue.RegisterHandler(TypeSendLaterEmail, HandleSendLaterEmailTask)

	registerCronJobs()
}

func registerCronJobs() {
	helloWorldTask, err := NewHelloWorldTask()
	if err != nil {
		log.Fatalf("failed to create 'hello world' cron task: %v", err)
	}
	queue.RegisterCronJob("@every 1m", helloWorldTask)
}
