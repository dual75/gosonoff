package somqtt

type NullMqttService struct {
	IncomingMessages chan *MqttIncomingMessage
}

func (s *NullMqttService) PublishToEventTopic(deviceid string, message interface{}) {
}

func (s *NullMqttService) PublishToActionTopic(deviceid string, message interface{}) {
}

func (s *NullMqttService) PublishToStatusTopic(deviceid string, status string) {
}

func (s *NullMqttService) SubscribeAll(deviceid string) (err error) {
	return
}

func (s *NullMqttService) UnsubscribeAll(deviceid string) (err error) {
	return
}

func (s *NullMqttService) GetIncomingMessages() chan *MqttIncomingMessage {
	return s.IncomingMessages
}
