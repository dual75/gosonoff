package somqtt

import (
	"encoding/json"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MessageHandler struct {
	Deviceid    string
	MqttService *MqttService
}

func (s *MessageHandler) ActionHandler(client mqtt.Client, message mqtt.Message) {
	var decoded map[string]interface{}
	err := json.Unmarshal(message.Payload(), &decoded)
	if err == nil {
		log.Printf("incoming mqtt message, %v", decoded)
		s.enqueue(&MqttIncomingMessage{CodeAction, (*s).Deviceid, &decoded})
	} else {
		log.Println(err)
	}
}

func (s *MessageHandler) SwitchHandler(client mqtt.Client, message mqtt.Message) {
	decoded := string(message.Payload()[:])
	s.enqueue(&MqttIncomingMessage{CodeSwitch, (*s).Deviceid, &decoded})
}

func (s *MessageHandler) enqueue(message *MqttIncomingMessage) {
	s.MqttService.GetIncomingMessages() <- message
}
