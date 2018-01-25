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
	"github.com/gorilla/mux"

	"github.com/dual75/gosonoff/sohttp"
	"github.com/dual75/gosonoff/sows"
)

// You can godoc variables

var (
	publisher somqtt.Publisher
	wsService *sows.WsService
)

// You can godoc functions

<<<<<<< HEAD
func runHttpServer(certfile *string, keyfile *string) (server *http.Server, err error) {
	router := mux.NewRouter()
	router.HandleFunc("/api/ws", wsService.ServeHTTP)
	handlers := sohttp.Handlers{sonoff.Config.Server.Addr, sonoff.Config.Server.Port, publisher, wsService}
	router.HandleFunc("/switch/{deviceid}/{status}", handlers.ServeSwitch).Methods("GET")
	router.HandleFunc("/switch/{deviceid}", handlers.ServeAction).Methods("POST")
	router.HandleFunc("/switch/{deviceid}", handlers.ServeStatus).Methods("GET")
	router.HandleFunc("/dispatch/device", handlers.ServeDevice).Methods("GET")
	server = &http.Server{
		Addr:           fmt.Sprintf("%v:%d", sonoff.Config.Server.Addr, sonoff.Config.Server.Port),
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
=======
func runHttpServer(certfile *string, keyfile *string, ch chan int) {
	r := mux.NewRouter()
	r.HandleFunc("/api/ws", wsService.ServeHTTP)
	server := sohttp.HTTPServer{sonoff.Config.Server.Addr, sonoff.Config.Server.Port, mqttService, wsService}
	r.HandleFunc("/switch/{deviceid}/{status}", server.ServeSwitch).Methods("GET")
	r.HandleFunc("/switch/{deviceid}", server.ServeAction).Methods("POST")
	r.HandleFunc("/switch/{deviceid}", server.ServeStatus).Methods("GET")
	r.HandleFunc("/dispatch/device", server.ServeDevice).Methods("GET")
	http.Handle("/", r)

	serveraddr := fmt.Sprintf("%v:%d", sonoff.Config.Server.Addr, sonoff.Config.Server.Port)
	err := http.ListenAndServeTLS(serveraddr, *certfile, *keyfile, nil)
	outcome := 0
	if err != nil {
		log.Println(err)
		outcome = 1
>>>>>>> bc360352886e08c46b3a57390c0afff96ffc7ee7
	}

	go func() {
		server.ListenAndServeTLS(*certfile, *keyfile)
		if err != nil {
			log.Println(err)
		}
	}()

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
				publisher.UnsubscribeAll(message.Deviceid)
			} else {
				log.Printf("message code %v, processed without errors\n", message.Code)
			}
		case <-signals:
			log.Println("caught interrupt signal")
			break
		}
	}
	return
}

func serve(certfile *string, keyfile *string) (err error) {
	publisher, err = somqtt.NewPublisher(sonoff.Config.Mqtt)
	checkErr(err)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	wsService = sows.NewWsService(publisher)
	server, err := runHttpServer(certfile, keyfile)

	if err != nil {
		log.Fatal(err)
	}

	err = selectEvents(publisher.GetIncomingMessages(), interrupt)
	if err == nil {
		// The request has a timeout, so create a context that is
		// canceled automatically when the timeout expires.
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		server.Shutdown(ctx)
		defer cancel()
	}

	return
}
