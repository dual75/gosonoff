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

type WsReply struct {
	Apikey   string                  `json:"apikey"`
	Error    int                     `json:"error"`
	Deviceid string                  `json:"deviceid"`
	Date     time.Time               `json:"date,omitempty"`
	Params   *map[string]interface{} `json:"params,omitempty"`
}

type WsRequest struct {
	Apikey   string                  `json:"apikey"`
	Deviceid string                  `json:"deviceid"`
	Action   string                  `json:"action"`
	Params   *map[string]interface{} `json:"params,omitempty"`
}

func NewWsReply(deviceid string) *WsReply {
	return &WsReply{
		Apikey:   sonoff.ApiKey,
		Error:    0,
		Deviceid: deviceid,
		Date:     time.Now(),
	}
}

func NewWsRequest(deviceid string, action string) *WsRequest {
	return &WsRequest{
		Apikey:   sonoff.ApiKey,
		Deviceid: deviceid,
		Action:   action,
	}
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (ws WsService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer ws.CleanClose(connection)
	for {
		var req WsRequest
		err := connection.ReadJSON(&req)
		if err != nil {
			log.Printf("read: %v\n", err)
			break
		}
		response, err := ws.dispatchRequest(&req, connection)
		if err != nil {
			log.Printf("dispatch: %v\n", err)
			continue
		}
		if response != nil {
			err = connection.WriteJSON(response)
			if err != nil {
				log.Printf("write: %v\n", err)
				break
			}
		}
	}
}

func (ws *WsService) dispatchRequest(req *WsRequest, conn *websocket.Conn) (*WsReply, error) {
	result := NewWsReply((*req).Deviceid)
	marshaled, err := json.Marshal(req)
	if err != nil {
		log.Printf("Error in hook marshaling %v", err)
	}

	go (*ws).mqttservice.PublishToEventTopic((*req).Deviceid, marshaled)
	log.Printf("ws action: %s, %v", (*req).Action, string(marshaled[:]))

	switch (*req).Action {
	case "date":
	case "update":
	case "register":
		ws.mux.Lock()
		defer ws.mux.Unlock()
		_, ok := ws.connections[(*req).Deviceid]
		if !ok {
			ws.connections[(*req).Deviceid] = conn
		}
		go (*ws).mqttservice.SubscribeAll((*req).Deviceid)
	case "query":
		(*result).Params = &map[string]interface{}{}
	default:
		result = nil
	}
	return result, err
}

func (ws *WsService) WriteTo(deviceId string, data interface{}) (err error) {
	ws.mux.Lock()
	defer ws.mux.Unlock()
	connection, ok := ws.connections[deviceId]
	if ok == false {
		err = fmt.Errorf("unknown device id: %v", deviceId)
	}
	if err == nil {
		err = connection.WriteJSON(data)
	}
	return
}

func (ws *WsService) Switch(deviceId string, flag string) (err error) {
	request := &WsRequest{
		Apikey:   sonoff.ApiKey,
		Action:   "update",
		Deviceid: deviceId,
		Params: &map[string]interface{}{
			"switch": flag,
		},
	}
	err = ws.WriteTo(deviceId, request)
	return
}
