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

func (s *SonoffMessageHandler) ActionHandler(client mqtt.Client, message mqtt.Message) {
	decoded := make(map[string]interface{})
	err := json.Unmarshal(message.Payload(), &decoded)
	if err == nil {
		s.enqueue(&MqttIncomingMessage{CodeAction, (*s).Deviceid, &decoded})
	} else {
		log.Println(err)
	}
}

func (s *SonoffMessageHandler) SwitchHandler(client mqtt.Client, message mqtt.Message) {
	decoded := string(message.Payload()[:])
	s.enqueue(&MqttIncomingMessage{CodeSwitch, (*s).Deviceid, &decoded})
}

func (s *SonoffMessageHandler) enqueue(message *MqttIncomingMessage) {
	(*s).MqttService.IncomingMessages <- message
}
