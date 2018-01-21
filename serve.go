package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dual75/gosonoff/somqtt"
	"github.com/gorilla/mux"

	"github.com/dual75/gosonoff/sohttp"
	"github.com/dual75/gosonoff/sows"
)

var mqttService *somqtt.MqttService

var wsService *sows.WsService

func runHttpServer(config sohttp.SonoffHttp, certfile *string, keyfile *string, ch chan int) {
	r := mux.NewRouter()
	r.HandleFunc("/api/ws", (*wsService).ServeHTTP)
	server := sohttp.HTTPServer{config.Addr, config.Port, mqttService}
	r.HandleFunc("/switch/{deviceid}/{status}", server.ServeSwitch)
	r.HandleFunc("/switch/{deviceid}/action", server.ServeAction)
	r.HandleFunc("/dispatch/device", server.ServeHTTP)
	http.Handle("/", r)

	serveraddr := fmt.Sprintf("%v:%d", config.Addr, config.Port)
	err := http.ListenAndServeTLS(serveraddr, *certfile, *keyfile, nil)
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
				err = fmt.Errorf("error outcome %v from runHttpServer", outcome)
			}
			break
		case message := <-mqttch:
			switch message.Code {
			case somqtt.CodeAction:
				log.Println("forwarding ", (*message).Message)
				err = wsService.WriteTo((*message).Deviceid, (*message).Message)
			case somqtt.CodeSwitch:
				flag := (*message).Message.(*string)
				err = wsService.Switch((*message).Deviceid, *flag)
			}

			if err != nil {
				log.Println(err)
				wsService.RemoveDeviceConnection((*message).Deviceid)
				mqttService.UnsubscribeAll((*message).Deviceid)
			} else {
				log.Printf("message code %v, processed without errors\n", message.Code)
			}
		}
	}
	return
}

func serve(certfile *string, keyfile *string) (err error) {
	mqttService, err = somqtt.NewMqttService(sonoffConfig.Mqtt)
	checkErr(err)

	wsService = sows.NewWsService(mqttService)
	serverChan := make(chan int)
	go runHttpServer(sonoffConfig.Server, certfile, keyfile, serverChan)
	err = selectEvents(serverChan, (*mqttService).IncomingMessages)
	return
}
