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

func checkError(token mqtt.Token) (err error) {
	token.Wait()
	if err = token.Error(); err != nil {
		err = fmt.Errorf("error in mqtt operation: %v", err)
	}
	return
}

func (s *MqttService) PublishToEventTopic(deviceid string, message interface{}) {
	if (*s).Client.IsConnected() {
		topic := fmt.Sprintf((*s).Config.Publishtopic, deviceid)
		checkError((*s).Client.Publish(topic, (*s).Config.Qos, false, message))
	} else {
		log.Printf("mqtt client not connected on trying to publish for device %v", deviceid)
	}
}

func (s *MqttService) PublishToActionTopic(deviceid string, message interface{}) {
	if (*s).Client.IsConnected() {
		topic := fmt.Sprintf((*s).Config.Actiontopic, deviceid)
		checkError((*s).Client.Publish(topic, (*s).Config.Qos, false, message))
		/*token.Wait()
		if err := token.Error(); err != nil {
			err = fmt.Errorf("errore in publish: %v", err)
		}
		*/
	} else {
		log.Printf("mqtt client not connected on trying to publish for device %v\n", deviceid)
	}
}

func (s *MqttService) Subscribe(deviceid string) (err error) {
	if (*s).Client.IsConnected() {
		log.Printf("now subscribing to topics for %v\n", deviceid)
		handler := &SonoffMessageHandler{deviceid, s}
		topic := fmt.Sprintf((*s).Config.Actiontopic, deviceid)
		err = s.subscribeTopic(topic, handler.ActionHandler)
		if err == nil {
			topic = fmt.Sprintf((*s).Config.Switchtopic, deviceid)
			err = (*s).subscribeTopic(topic, handler.SwitchHandler)
		}
		if err != nil {
			s.Unsubscribe(deviceid)
		}
	} else {
		log.Printf("mqtt client not connected, while trying to subscribe for device %v", deviceid)
	}
	return
}

func (s *MqttService) subscribeTopic(topic string, handler mqtt.MessageHandler) (err error) {
	err = checkError((*s).Client.Subscribe(topic, (*s).Config.Qos, handler))
	if err == nil {
		log.Printf("subscribe to %v ok\n", topic)
	}
	return
}

func (s *MqttService) Unsubscribe(deviceid string) (err error) {
	log.Printf("now unsubscribing from all topics %v\n", deviceid)
	for _, topic := range []string{(*s).Config.Actiontopic, (*s).Config.Switchtopic} {
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
	err = checkError((*s).Client.Unsubscribe(topic))
	if err == nil {
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
