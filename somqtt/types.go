package somqtt

type SonoffMqtt struct {
	Enabled     bool
	Eventtopic  string
	Actiontopic string
	Switchtopic string
	Statustopic string
	Url         string
	Qos         byte
}

type MqttMessageCode int

type MqttIncomingMessage struct {
	Code     MqttMessageCode
	Deviceid string
	Message  interface{}
}
