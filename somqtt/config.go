package somqtt

type SonoffMqtt struct {
	Enabled        bool
	Publishtopic   string
	Subscribetopic string
	Switchtopic    string
	Url            string
	Qos            byte
}
