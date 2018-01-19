package somqtt

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttIncomingMessage struct {
	Deviceid string
	Message  *map[string]interface{}
}

type MqttService struct {
	Config           SonoffMqtt
	Client           mqtt.Client
	IncomingMessages chan *MqttIncomingMessage
}

func (s *MqttService) Publish(deviceid string, message interface{}) {
	if (*s).Client.IsConnected() {
		topic := fmt.Sprintf((*s).Config.Publishtopic, deviceid)
		token := (*s).Client.Publish(topic, (*s).Config.Qos, false, message)
		token.Wait()
		if err := token.Error(); err != nil {
			err = fmt.Errorf("Errore in publish: %v", err)
		}
	} else {
		log.Printf("mqtt client not connected, dispatching for device %v", deviceid)
	}
}

func (s *MqttService) Subscribe(deviceid string) (err error) {
	log.Printf("now subscribing to topic %v\n", deviceid)
	topic := fmt.Sprintf((*s).Config.Subscribetopic, deviceid)
	handler := &SonoffMessageHandler{deviceid, s}
	token := (*s).Client.Subscribe(topic, (*s).Config.Qos, handler.MessageHandler)
	token.Wait()
	if err = token.Error(); err != nil {
		err = fmt.Errorf("Error %v in Subscribe topic %v", err, topic)
	}
	log.Printf("subscription to  %v ok\n", topic)
	return
}

func (s *MqttService) Unsubscribe(deviceid string) (err error) {
	log.Printf("now unsubscribing to topic %v\n", deviceid)
	topic := fmt.Sprintf((*s).Config.Subscribetopic, deviceid)
	token := (*s).Client.Unsubscribe(topic)
	token.Wait()
	if err = token.Error(); err != nil {
		err = fmt.Errorf("Error %v in Unsubscribe topic %v", err, topic)
	}
	log.Printf("unsubscription to  %v ok\n", topic)
	return
}

func NewMqttService(config SonoffMqtt) (result *MqttService, err error) {
	result = &MqttService{}
	copt := mqtt.NewClientOptions()
	copt.AddBroker(config.Url)
	copt.SetCleanSession(true)
	copt.SetAutoReconnect(true)
	result = &MqttService{config, mqtt.NewClient(copt), make(chan *MqttIncomingMessage)}
	token := result.Client.Connect()
	token.Wait()
	if err = token.Error(); err != nil {
		log.Fatal(err)
	}
	return
}
