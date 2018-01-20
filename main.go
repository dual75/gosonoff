package main

import (
	"io/ioutil"
	"log"

	"github.com/dual75/gosonoff/sohttp"
	"github.com/dual75/gosonoff/somqtt"
	"github.com/dual75/gosonoff/sows"
	"github.com/go-yaml/yaml"
)

type SonoffConfig struct {
	Server sohttp.SonoffHttp
	Mqtt   somqtt.SonoffMqtt
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
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

	err = selectEvents(serverChan, (*mqttService).IncomingMessages)
	checkErr(err)
}
