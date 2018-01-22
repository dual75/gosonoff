package sows

import (
	"log"
	"net/http"
	"sync"

	"github.com/dual75/gosonoff/somqtt"
	"github.com/dual75/gosonoff/sonoff"
	"github.com/gorilla/websocket"
)

type WsService struct {
	devices     map[string]*ConnectedDevice
	mqttservice *somqtt.MqttService
	upgrader    websocket.Upgrader
	dataMux     sync.Mutex
}

func NewWsService(mqttservice *somqtt.MqttService) (result *WsService) {
	result = &WsService{
		mqttservice: mqttservice,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	return
}

func (ws *WsService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	device, err := RegisterConnectedDevice(conn, (*ws).mqttservice)
	if err == nil {
		defer ws.removeConnectedDevice(device)

		(*ws).devices[device.Deviceid] = device
		device.ServeForever()
	} else {
		defer conn.Close()
	}
}

func (ws *WsService) removeConnectedDevice(connectedDevice *ConnectedDevice) {
	(*ws).dataMux.Lock()
	defer (*ws).dataMux.Unlock()

	err := connectedDevice.CloseConnection()
	if err != nil {
		log.Println("error in connecteddevice.CloseConnection():", err)
	}
	delete((*ws).devices, (*connectedDevice).Deviceid)
}

func (ws *WsService) WriteTo(deviceId string, data interface{}, callback ReplyCallback) (err error) {
	if device, ok := ws.devices[deviceId]; ok {
		err = device.Write(data, callback)
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
	successCallback := func(device *ConnectedDevice, message *WsMessage) {
		go (*ws).mqttservice.PublishToStatusTopic(deviceId, (*device).Status)
	}
	err = ws.WriteTo(deviceId, request, successCallback)
	return
}

func (ws *WsService) Status(deviceId string) (err error) {
	request := &WsMessage{
		Apikey:   sonoff.ApiKey,
		Action:   "status",
		Deviceid: deviceId,
		Params:   []string{"switch"},
	}
	successCallback := func(device *ConnectedDevice, message *WsMessage) {
		params := (*message).Params.(map[string]interface{})
		(*device).Status = params["switch"].(string)
		go (*ws).mqttservice.PublishToStatusTopic(deviceId, (*device).Status)
	}
	err = ws.WriteTo(deviceId, request, successCallback)
	return
}

func (ws *WsService) DiscardDevice(deviceId string) {
	if device, ok := (*ws).devices[deviceId]; ok {
		ws.removeConnectedDevice(device)
	}
}
