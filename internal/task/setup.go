package task

import (
	"mygoframe/pkg/queue"
)

func Setup() {
	queue.RegisterHandler(TypeWelcomeEmail, HandleWelcomeEmailTask)
	queue.RegisterHandler(TypeSendLaterEmail, HandleSendLaterEmailTask)

	queue.RegisterCronJob("@every 1m", NewHelloWorldTask())
}
