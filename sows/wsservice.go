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

type ConnectedDevice struct {
	Deviceid   string
	Devicetype string
	Romversion string
	Model      string
	Connection *websocket.Conn
	Status     string
}

type WsService struct {
	devices     map[string]*ConnectedDevice
	mqttservice *somqtt.MqttService
	mux         sync.Mutex
}

func (ws *WsService) RemoveDeviceConnection(deviceid string) {
	if conn, ok := (*ws).devices[deviceid]; ok {
		(*conn).Connection.Close()
		delete((*ws).devices, deviceid)
	}
}

func (ws *WsService) CleanClose(conn *websocket.Conn) {
	(*ws).mux.Lock()
	err := (*conn).Close()
	if err != nil {
		log.Printf("Error while closing device: %v", err)
	}
	if err == nil {
		for key, value := range (*ws).devices {
			if conn == value.Connection {
				delete((*ws).devices, key)
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
	device, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer ws.CleanClose(device)
	for {
		var req WsMessage
		err := device.ReadJSON(&req)
		if err != nil {
			log.Printf("read: %v\n", err)
			break
		}
		response, err := ws.dispatchRequest(&req, device)
		if err != nil {
			log.Printf("dispatch: %v\n", err)
			continue
		}
		if response != nil {
			err = device.WriteJSON(response)
			if err != nil {
				log.Printf("write: %v\n", err)
				break
			}
		}
	}
}

func (ws *WsService) dispatchRequest(req *WsMessage, conn *websocket.Conn) (*WsMessage, error) {
	result := NewWsMessage((*req).Deviceid)
	(*result).Error = 0
	marshaled, err := json.Marshal(req)
	if err != nil {
		log.Printf("Error in hook marshaling %v", err)
	}

	go (*ws).mqttservice.PublishToEventTopic((*req).Deviceid, marshaled)
	log.Printf("ws action: %s, %v", (*req).Action, string(marshaled[:]))

	switch (*req).Action {
	case "date":
		(*result).Date = time.Now()
	case "update":
		params := (*req).Params.(map[string]interface{})
		status, ok := params["switch"]
		if ok {
			ws.devices[(*req).Deviceid].Status = status.(string)
			go (*ws).mqttservice.PublishToStatusTopic((*req).Deviceid, status.(string))
		}
	case "register":
		ws.mux.Lock()
		defer ws.mux.Unlock()
		_, ok := ws.devices[(*req).Deviceid]
		if !ok {
			devicetype := GetDeviceType((*req).Deviceid)
			ws.devices[(*req).Deviceid] = &ConnectedDevice{
				(*req).Deviceid,
				devicetype,
				(*req).Romversion,
				(*req).Model,
				conn,
				"unknown",
			}
			go (*ws).mqttservice.PublishToStatusTopic((*req).Deviceid, "unknown")
			go (*ws).mqttservice.SubscribeAll((*req).Deviceid)
		}
	case "query":
		(*result).Params = map[string]string{}
	default:
		result = nil
	}
	return result, err
}

func GetDeviceType(deviceid string) (result string) {
	prefix := deviceid
	switch prefix {
	case "01":
		result = "switch"
	case "02":
		result = "light"
	case "03":
		result = "sensor"
	default:
		result = "unknown"
	}
	return
}

func (ws *WsService) WriteTo(deviceId string, data interface{}) (err error) {
	ws.mux.Lock()
	defer ws.mux.Unlock()
	device, ok := ws.devices[deviceId]
	if ok == false {
		err = fmt.Errorf("unknown device id: %v", deviceId)
	}
	if err == nil {
		err = device.Connection.WriteJSON(data)
	}
	return
}

func (ws *WsService) Switch(deviceId string, flag string) (err error) {
	request := &WsMessage{
		Apikey:   sonoff.ApiKey,
		Action:   "update",
		Deviceid: deviceId,
		Params: map[string]string{
			"switch": flag,
		},
	}
	err = ws.WriteTo(deviceId, request)
	return
}

func NewWsService(mqttservice *somqtt.MqttService) (result *WsService) {
	result = &WsService{make(map[string]*ConnectedDevice), mqttservice, sync.Mutex{}}
	return
}
