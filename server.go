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

func runHttpServer(config sohttp.SonoffHttp, ch chan int) {
	r := mux.NewRouter()
	r.HandleFunc("/api/ws", (*wsService).ServeHTTP)
	server := sohttp.HTTPServer{config.Addr, config.Port, mqttService}
	r.HandleFunc("/switch/{deviceid}/{status}", server.ServeSwitch)
	r.HandleFunc("/switch/{deviceid}/action", server.ServeMessage)
	r.HandleFunc("/dispatch/device", server.ServeHTTP)
	http.Handle("/", r)

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
