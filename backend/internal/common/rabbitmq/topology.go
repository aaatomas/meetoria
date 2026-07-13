package rabbitmq

const (
	ExchangeType = "topic"

	BackendExchangeName = "meetoria-backend"
	BackendQueueName    = "meetoria-backend"

	SMSWorkerExchangeName   = "meetoria-sms-worker"
	SMSWorkerQueueName      = "meetoria-sms-worker"
	EmailWorkerExchangeName = "meetoria-email-worker"
	EmailWorkerQueueName    = "meetoria-email-worker"

	RoutingNotificationSMS           = "notification.sms"
	RoutingNotificationEmail           = "notification.email"
	RoutingNotificationSMSProcessing   = "notification.sms.processing"
	RoutingNotificationSMSSent         = "notification.sms.sent"
	RoutingNotificationSMSFailed       = "notification.sms.failed"
	RoutingNotificationEmailProcessing = "notification.email.processing"
	RoutingNotificationEmailSent       = "notification.email.sent"
	RoutingNotificationEmailFailed     = "notification.email.failed"
)
