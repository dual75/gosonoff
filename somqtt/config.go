package somqtt

type SonoffMqtt struct {
	Enabled      bool
	Publishtopic string
	Actiontopic  string
	Switchtopic  string
	Url          string
	Qos          byte
}
