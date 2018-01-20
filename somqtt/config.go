package somqtt

type SonoffMqtt struct {
	Enabled     bool
	Eventtopic  string
	Actiontopic string
	Switchtopic string
	Url         string
	Qos         byte
}
