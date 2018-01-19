package sows

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dual75/gosonoff/somqtt"
	"github.com/dual75/gosonoff/sonoff"
	"github.com/gorilla/websocket"
)

const (
	WsReplyOk = 0
	WsReplyKo = 1
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsService struct {
	connections map[string]*websocket.Conn
	mqttservice *somqtt.MqttService
	mux         sync.Mutex
}

func NewWsService(mqttservice *somqtt.MqttService) (result *WsService) {
	result = &WsService{make(map[string]*websocket.Conn), mqttservice, sync.Mutex{}}
	return
}

func (ws *WsService) RemoveDeviceConnection(deviceid string) {
	if conn, ok := (*ws).connections[deviceid]; ok {
		(*conn).Close()
		delete((*ws).connections, deviceid)
	}
}

func (ws *WsService) CleanClose(conn *websocket.Conn) {
	(*ws).mux.Lock()
	err := (*conn).Close()
	if err != nil {
		log.Printf("Error while closing connection: %v", err)
	}
	if err == nil {
		for key, value := range (*ws).connections {
			if conn == value {
				delete((*ws).connections, key)
			}
		}
	}
	(*ws).mux.Unlock()
}

func (ws *WsService) Handler(w http.ResponseWriter, r *http.Request) {
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer ws.CleanClose(connection)

	for {
		var req interface{}
		err := connection.ReadJSON(&req)
		if err != nil {
			log.Println("read:", err)
			break
		}

		mapRequest := req.(map[string]interface{})
		response, err := ws.dispatchRequest(&mapRequest, connection)
		if err != nil {
			log.Println("dispatch:", err)
			continue
		}

		w.Header().Set("Content-Type", sonoff.ContentType)
		err = connection.WriteJSON(response)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func (ws *WsService) dispatchRequest(req *map[string]interface{}, conn *websocket.Conn) (*map[string]interface{}, error) {
	deviceid, action := (*req)["deviceid"].(string), (*req)["action"].(string)

	result := make(map[string]interface{})
	result["deviceid"] = deviceid
	result["error"] = WsReplyOk
	result["apiKey"] = sonoff.ApiKey

	marshaled, err := json.Marshal(req)
	if err != nil {
		log.Printf("Error in hook marshaling %v", err)
	}

	switch action {
	case "date":
		result["date"] = time.Now()
	case "update":
		go (*ws).mqttservice.Publish(deviceid, marshaled)
	case "register":
		ws.mux.Lock()
		defer ws.mux.Unlock()
		_, ok := ws.connections[deviceid]
		if !ok {
			ws.connections[deviceid] = conn
		}
		go (*ws).mqttservice.Subscribe(deviceid)
	case "query":
		result["params"] = make(map[string]interface{})
	default:
		result["error"] = WsReplyKo
	}
	return &result, err
}

func (ws *WsService) WriteTo(deviceId string, data *map[string]interface{}) (err error) {
	ws.mux.Lock()
	defer ws.mux.Unlock()

	connection, ok := ws.connections[deviceId]
	if ok == false {
		err = fmt.Errorf("Unknown device id: %v", deviceId)
	}
	if err == nil {
		err = connection.WriteJSON(data)
	}
	return
}
