package somqtt

type MqttMessageCode int

type MqttIncomingMessage struct {
	Code     MqttMessageCode
	Deviceid string
	Message  interface{}
}
