package somqtt

import (
	"encoding/json"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type SonoffMessageHandler struct {
	Deviceid    string
	MqttService *MqttService
}

func (s *SonoffMessageHandler) MessageHandler(client mqtt.Client, message mqtt.Message) {
	decoded := make(map[string]interface{})
	err := json.Unmarshal(message.Payload(), &decoded)
	if err == nil {
		(*s).MqttService.IncomingMessages <- &MqttIncomingMessage{(*s).Deviceid, &decoded}
	} else {
		log.Println(err)
	}
}
