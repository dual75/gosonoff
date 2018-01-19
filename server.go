package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dual75/gosonoff/somqtt"

	"github.com/go-yaml/yaml"

	"github.com/dual75/gosonoff/sohttp"
	"github.com/dual75/gosonoff/sows"
)

var mqttService *somqtt.MqttService

var wsService *sows.WsService

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func runHttpServer(config sohttp.SonoffHttp, ch chan int) {
	serveraddr := fmt.Sprintf("%v:%d", config.Addr, config.Port)
	httpServer := sohttp.HTTPServer{config.Addr, config.Port, config.Wsport}
	err := http.ListenAndServeTLS(serveraddr, "server.cert", "server.key", httpServer)
	outcome := 0
	if err != nil {
		log.Println(err)
		outcome = 1
	}
	ch <- outcome
}

func runWsServer(config sohttp.SonoffHttp, ch chan int) {
	serveraddr := fmt.Sprintf("%v:%d", config.Addr, config.Wsport)
	err := http.ListenAndServeTLS(serveraddr, "server.cert", "server.key", *wsService)
	outcome := 0
	if err != nil {
		log.Println(err)
		outcome = 1
	}
	ch <- outcome
}

func selectEvents(serverch chan int, wschan chan int, mqttch chan *somqtt.MqttIncomingMessage) (err error) {
	for {
		select {
		case outcome := <-serverch:
			if outcome == 1 {
				err = fmt.Errorf("Error, outcome %v from runServer", outcome)
			}
			break
		case outcome := <-wschan:
			if outcome == 1 {
				err = fmt.Errorf("Error, outcome %v from wsService", outcome)
			}
			break
		case message := <-mqttch:
			switch message.Code {
			case somqtt.CodeAction:
				err = wsService.WriteTo((*message).Deviceid, (*message).Message)
			case somqtt.CodeSwitch:
				flag := (*message).Message.(*string)
				err = wsService.Switch((*message).Deviceid, *flag)
			}

			if err != nil {
				log.Println(err)
				wsService.RemoveDeviceConnection((*message).Deviceid)
				mqttService.Unsubscribe((*message).Deviceid)
			} else {
				log.Printf("message code %v, processed without errors\n", message.Code)
			}
		}
	}
	return
}

type SonoffConfig struct {
	Server sohttp.SonoffHttp
	Mqtt   somqtt.SonoffMqtt
}

func main() {
	data, err := ioutil.ReadFile("config.yml")
	checkErr(err)

	config := SonoffConfig{}
	err = yaml.Unmarshal(data, &config)
	checkErr(err)

	mqttService, err = somqtt.NewMqttService(config.Mqtt)
	checkErr(err)

	wsService = sows.NewWsService(mqttService)

	serverChan := make(chan int)
	go runHttpServer(config.Server, serverChan)

	wsChan := make(chan int)
	go runWsServer(config.Server, wsChan)

	err = selectEvents(serverChan, wsChan, (*mqttService).IncomingMessages)
	checkErr(err)
}
