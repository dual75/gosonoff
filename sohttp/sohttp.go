package sohttp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dual75/gosonoff/somqtt"

	"github.com/dual75/gosonoff/sonoff"
	"github.com/dual75/gosonoff/sows"
	"github.com/gorilla/mux"
)

type WebSocketConfig struct {
	Error  int    `json:"error"`
	Reason string `json:"reason"`
	IP     string `json:"IP"`
	Port   int    `json:"port"`
}

type HTTPServer struct {
	Ip          string
	Port        int
	MqttService *somqtt.MqttService
}

// ServeHTTP Single handle func
func (server HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wsConfig := WebSocketConfig{0, "ok", server.Ip, server.Port}
	log.Printf("request: URI = %v", r.RequestURI)
	bytes, err := ioutil.ReadAll(r.Body)
	if err == nil {
		log.Printf("request: body = %v", string(bytes))
	} else {
		log.Println(err)
	}
	w.Header().Set("Content-Type", sonoff.ContentType)
	json.NewEncoder(w).Encode(wsConfig)
}

func (server HTTPServer) ServeSwitch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceid, status := vars["deviceid"], vars["status"]

	switch status {
	case "on", "off":
		request := &sows.WsRequest{
			Apikey:   sonoff.ApiKey,
			Action:   "update",
			Deviceid: deviceid,
			Params: &map[string]interface{}{
				"switch": status,
			},
		}
		encoded, err := json.Marshal(request)
		if err == nil {
			go (*server.MqttService).PublishToActionTopic(deviceid, encoded)
			w.WriteHeader(http.StatusOK)
		} else {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (server HTTPServer) ServeMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceid := vars["deviceid"]

	bytes, err := ioutil.ReadAll(r.Body)
	if err == nil {
		go (*server.MqttService).PublishToActionTopic(deviceid, bytes)
		w.WriteHeader(http.StatusOK)
	} else {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
	}
}
