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

func runServer(config sohttp.SonoffHttp, ch chan int) {
	serveraddr := fmt.Sprintf("%v:%d", config.Addr, config.Port)
	err := http.ListenAndServeTLS(serveraddr, "server.cert", "server.key", nil)
	outcome := 0
	if err != nil {
		log.Println(err)
		outcome = 1
	}
	ch <- outcome
}

func selectEvents(serverch chan int, mqttch chan *somqtt.MqttIncomingMessage) (err error) {
	for {
		select {
		case outcome := <-serverch:
			if outcome == 1 {
				err = fmt.Errorf("Error, outcome %v from runServer", outcome)
			}
			break
		case message := <-mqttch:
			err = wsService.WriteTo((*message).Deviceid, (*message).Message)
			if err != nil {
				log.Println(err)
				wsService.RemoveDeviceConnection((*message).Deviceid)
				mqttService.Unsubscribe((*message).Deviceid)
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
	http.HandleFunc("/ws", wsService.Handler)
	http.HandleFunc("/", sohttp.ServeHTTP)

	serverChan := make(chan int)
	go runServer(config.Server, serverChan)
	err = selectEvents(serverChan, (*mqttService).IncomingMessages)

	checkErr(err)
}
