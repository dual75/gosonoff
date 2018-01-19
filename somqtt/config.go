package somqtt

type SonoffMqtt struct {
	Enabled        bool
	Publishtopic   string
	Subscribetopic string
	Url            string
	Qos            byte
}
