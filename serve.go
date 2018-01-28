package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/dual75/gosonoff/somqtt"
	"github.com/dual75/gosonoff/sonoff"

	"github.com/dual75/gosonoff/sohttp"
	"github.com/dual75/gosonoff/sows"

"github.com/gorilla/mux"
)

// You can godoc variables

var (
	mqttService somqtt.MqttService
	wsService   *sows.WsService
)

// You can godoc functions

func makeHttpServer(certfile *string, keyfile *string) (server *http.Server, err error) {
	handlers := sohttp.Handlers{sonoff.Config.Server.Addr, sonoff.Config.Server.Port, mqttService, wsService}
	router := mux.NewRouter()
	router.HandleFunc("/switch/{deviceid}/{status}", handlers.HandleSwitch)
	router.HandleFunc("/switch/{deviceid}", handlers.HandleAction)
	router.HandleFunc("/dispatch/device", handlers.HandleDevice)
	router.Handle("/api/ws", wsService)
	server = &http.Server{
		Addr:           fmt.Sprintf("%v:%d", sonoff.Config.Server.Addr, sonoff.Config.Server.Port),
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return server, err
}

func selectEvents(mqttch <-chan *somqtt.MqttIncomingMessage, signals <-chan os.Signal) (err error) {
	for {
		select {
		case message := <-mqttch:
			switch message.Code {
			case somqtt.CodeAction:
				log.Println("forwarding ", message.Message)
				err = wsService.WriteTo(message.Deviceid, message.Message, nil)
			case somqtt.CodeSwitch:
				flag := message.Message.(*string)
				err = wsService.Switch(message.Deviceid, *flag)
			}
			if err != nil {
				log.Println(err)
				wsService.DiscardDevice(message.Deviceid)
				mqttService.UnsubscribeAll(message.Deviceid)
			} else {
				log.Printf("message code %v, processed without errors\n", message.Code)
			}
		case <-signals:
			log.Println("caught interrupt signal")
			return
		}
	}
	return
}

func serve(certfile *string, keyfile *string) (err error) {
	mqttService, err = somqtt.NewMqttService(sonoff.Config.Mqtt)
	checkErr(err)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	wsService = sows.NewWsService(mqttService)
	server, err := makeHttpServer(certfile, keyfile)

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err := server.ListenAndServeTLS(*certfile, *keyfile)
		if err != nil {
			log.Println(err)
		}
	}()

	err = selectEvents(mqttService.GetIncomingMessages(), interrupt)
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
		server.Shutdown(ctx)
		defer cancel()
	}

	return
}
