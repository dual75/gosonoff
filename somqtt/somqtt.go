package somqtt

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	CodeAction = iota
	CodeSwitch = iota
)

type MqttMessageCode int

type MqttIncomingMessage struct {
	Code     MqttMessageCode
	Deviceid string
	Message  interface{}
}

type MqttService struct {
	Config           SonoffMqtt
	Client           mqtt.Client
	IncomingMessages chan *MqttIncomingMessage
}

func (s *MqttService) Publish(deviceid string, message interface{}) {
	if (*s).Client.IsConnected() {
		subTopic := fmt.Sprintf((*s).Config.Publishtopic, deviceid)
		log.Printf("publishing to subTopic %s, value: %v\n", message)
		token := (*s).Client.Publish(subTopic, (*s).Config.Qos, false, message)
		token.Wait()
		if err := token.Error(); err != nil {
			err = fmt.Errorf("Errore in publish: %v", err)
		}
	} else {
		log.Printf("mqtt client not connected, dispatching for device %v", deviceid)
	}
}

func (s *MqttService) Subscribe(deviceid string) (err error) {
	log.Printf("now subscribing to topics for %v\n", deviceid)
	handler := &SonoffMessageHandler{deviceid, s}

	topic := fmt.Sprintf((*s).Config.Subscribetopic, deviceid)
	err = s.subscribeTopic(topic, handler.ActionHandler)
	if err == nil {
		topic = fmt.Sprintf((*s).Config.Switchtopic, deviceid)
		err = (*s).subscribeTopic(topic, handler.SwitchHandler)
	}
	if err != nil {
		s.Unsubscribe(deviceid)
	}
	return
}

func (s *MqttService) subscribeTopic(topic string, handler mqtt.MessageHandler) (err error) {
	token := (*s).Client.Subscribe(topic, (*s).Config.Qos, handler)
	token.Wait()
	if err = token.Error(); err != nil {
		err = fmt.Errorf("Error %v in Subscribe topic %v", err, topic)
	} else {
		log.Printf("subscription to  %v ok\n", topic)
	}
	return
}

func (s *MqttService) Unsubscribe(deviceid string) (err error) {
	log.Printf("now unsubscribing from all topics %v\n", deviceid)
	for _, topic := range []string{(*s).Config.Subscribetopic, (*s).Config.Switchtopic} {
		complete := fmt.Sprintf(topic, deviceid)
		err = s.unsubscribeTopic(deviceid, complete)
		if err != nil {
			log.Println(err)
		}
	}
	return
}

func (s *MqttService) unsubscribeTopic(deviceid string, topic string) (err error) {
	log.Printf("unsubscribe from topic %v\n", topic)
	token := (*s).Client.Unsubscribe(topic)
	token.Wait()
	if err = token.Error(); err != nil {
		err = fmt.Errorf("Error %v in Unsubscribe topic %v", err, topic)
	} else {
		log.Printf("unsubscription to  %v ok\n", topic)
	}
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
