package somqtt

type MqttService interface {
	PublishToEventTopic(deviceid string, message interface{})

	PublishToActionTopic(deviceid string, message interface{})

	PublishToStatusTopic(deviceid string, status string)

	SubscribeAll(deviceid string) (err error)

	UnsubscribeAll(deviceid string) (err error)

	GetIncomingMessages() chan *MqttIncomingMessage
}

