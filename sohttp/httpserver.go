package sohttp

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dual75/gosonoff/somqtt"

	"github.com/dual75/gosonoff/sonoff"
	"github.com/dual75/gosonoff/sows"
	"github.com/gorilla/mux"
)

type HTTPServer struct {
	Ip          string
	Port        int
	MqttService *somqtt.MqttService
	WsService   *sows.WsService
}

type WebSocketConfig struct {
	Error  int    `json:"error"`
	Reason string `json:"reason"`
	IP     string `json:"IP"`
	Port   int    `json:"port"`
}

// ServeHTTP Single handle func
func (server *HTTPServer) ServeDevice(w http.ResponseWriter, r *http.Request) {
	wsConfig := WebSocketConfig{0, "ok", server.Ip, server.Port}
	log.Printf("request: URI = %v", r.RequestURI)
	bytes, err := ioutil.ReadAll(r.Body)
	if err == nil {
		log.Printf("request: body = %v", string(bytes))
	} else {
		log.Println(err)
	}
	w.Header().Set("Content-Type", sonoff.ContentType)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(wsConfig)
}

func (server *HTTPServer) ServeSwitch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceid, status := vars["deviceid"], vars["status"]

	switch status {
	case "on", "off":
		request := &sows.WsMessage{
			Apikey:   sonoff.ApiKey,
			Action:   "update",
			Deviceid: deviceid,
			Params: map[string]string{
				"switch": status,
			},
		}
		encoded, err := json.Marshal(request)
		if err == nil {
			go (*server.MqttService).PublishToActionTopic(deviceid, encoded)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(MSGOk))
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(MSGError))
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(MSGError))
	}
}

func (server *HTTPServer) ServeAction(w http.ResponseWriter, r *http.Request) {
	deviceid := mux.Vars(r)["deviceid"]

	bytes, err := ioutil.ReadAll(r.Body)
	if err == nil {
		go (*server.MqttService).PublishToActionTopic(deviceid, bytes)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(MSGOk))
	} else {
		log.Println(err)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(MSGError))
	}
}

func (server *HTTPServer) ServeStatus(w http.ResponseWriter, r *http.Request) {
	deviceid := mux.Vars(r)["deviceid"]

	device, err := server.WsService.DeviceById(deviceid)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(device)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
	}
}
