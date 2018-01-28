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

type WebSocketConfig struct {
	Error  int    `json:"error"`
	Reason string `json:"reason"`
	IP     string `json:"IP"`
	Port   int    `json:"port"`
}

type Handlers struct {
	Ip          string
	Port        int
	MqttService somqtt.MqttService
	WsService   *sows.WsService
}

// ServeHTTP Single handle func
func (server *Handlers) HandleDevice(w http.ResponseWriter, r *http.Request) {
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

func (server *Handlers) HandleSwitch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceid, switchStatus := vars["deviceid"], vars["status"]
	status, message := http.StatusOK, MSGOk
	switch switchStatus {
	case "on", "off":
		request := &sows.WsMessage{
			Apikey:   sonoff.ApiKey,
			Action:   "update",
			Deviceid: deviceid,
			Params: map[string]string{
				"switch": switchStatus,
			},
		}
		encoded, err := json.Marshal(request)
		if err == nil {
			go server.MqttService.PublishToActionTopic(deviceid, encoded)
		} else {
			status, message = http.StatusInternalServerError, err.Error()
		}
	default:
		status, message = http.StatusBadRequest, MSGError
	}
	w.WriteHeader(status)
	w.Write([]byte(message))
}

func (server *Handlers) HandleAction(w http.ResponseWriter, r *http.Request) {
	deviceid := mux.Vars(r)["deviceid"]
	status, message := http.StatusOK, MSGOk
	if r.Method == http.MethodPost {
		bytes, err := ioutil.ReadAll(r.Body)
		if err == nil {
			go server.MqttService.PublishToActionTopic(deviceid, bytes)
		} else {
			status, message = http.StatusInternalServerError, err.Error()
		}
	} else {
		status, message = http.StatusMethodNotAllowed, MSGError
	}
	w.WriteHeader(status)
	w.Write([]byte(message))
}

func (server *Handlers) HandleStatus(w http.ResponseWriter, r *http.Request) {
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
