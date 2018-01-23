package sonoff

// SonoffConfig holds configuration from yaml file

type SonoffHttp struct {
	Addr string
	Port int
}

type SonoffMqtt struct {
	Enabled     bool
	Eventtopic  string
	Actiontopic string
	Switchtopic string
	Statustopic string
	Url         string
	Qos         byte
}

type SonoffConfig struct {
	Server SonoffHttp
	Mqtt   SonoffMqtt
}
