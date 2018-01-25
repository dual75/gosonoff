package somqtt

import (
	"log"

	"github.com/dual75/gosonoff/sonoff"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func NewMqttService(config sonoff.SonoffMqtt) (result Publisher, err error) {
	if config.Enabled {
		mqttservice := &MqttService{}
		copt := mqtt.NewClientOptions()
		copt.AddBroker(config.Url)
		copt.SetCleanSession(true)
		copt.SetAutoReconnect(true)
		result = &MqttService{config, mqtt.NewClient(copt), make(chan *MqttIncomingMessage)}
		token := mqttservice.Client.Connect()
		token.Wait()
		if err = token.Error(); err != nil {
			log.Fatal(err)
		}
		result = mqttservice
	} else {
		result = &NullMqttService{make(chan *MqttIncomingMessage)}
	}
	return
}
