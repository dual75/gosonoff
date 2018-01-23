// sows

package sows

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/dual75/gosonoff/somqtt"
	"github.com/dual75/gosonoff/sonoff"
	"github.com/gorilla/websocket"
)

// You can godoc types

// WsService is a service object which acts as a facade for WebSocket communications
type WsService struct {
	devices     map[string]*ConnectedDevice
	mqttservice *somqtt.MqttService
	upgrader    websocket.Upgrader
	dataMux     sync.Mutex
}

// You can godoc functions

// NewWsService creates a new WsService
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

// ServeHTTP is a handler function for http server
// Connection is upgraded to websocket and a new ConnectedDevice is created
func (ws *WsService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	device, err := RegisterConnectedDevice(conn, ws.mqttservice)
	if err == nil {
		defer ws.removeConnectedDevice(device)

		ws.devices[device.Deviceid] = device
		device.ServeForever()
	} else {
		defer conn.Close()
	}
}

// removeConnectedDevice removes an existing ConnectedDevice an request for connection to be closed
func (ws *WsService) removeConnectedDevice(connectedDevice *ConnectedDevice) {
	ws.dataMux.Lock()
	defer ws.dataMux.Unlock()

	err := connectedDevice.CloseConnection()
	if err != nil {
		log.Println("error in connecteddevice.CloseConnection():", err)
	}
	delete(ws.devices, (*connectedDevice).Deviceid)
}

// WriteTo writes an interface{} argument (see json.Marshal documentation) to the ConnectedDevice via websocket
func (ws *WsService) WriteTo(deviceId string, data interface{}, callback ReplyCallback) (err error) {
	if device, ok := ws.devices[deviceId]; ok {
		err = device.Write(data, callback)
	} else {
		log.Println("error in WriteTo, unknown device", deviceId)
	}
	return
}

// Switch crafts a switch update request and send it to the ConnectedDevice
func (ws *WsService) Switch(deviceId string, flag string) (err error) {
	request := &WsMessage{
		Apikey:   sonoff.ApiKey,
		Action:   "update",
		Deviceid: deviceId,
		Params: map[string]string{
			"switch": flag,
		},
	}
	successCallback := func(device *ConnectedDevice, message *WsMessage) (err error) {
		go ws.mqttservice.PublishToStatusTopic(deviceId, (*device).Status)
		return
	}
	err = ws.WriteTo(deviceId, request, successCallback)
	return
}

// Status crafts a switch status request and send it to the ConnectedDevice
func (ws *WsService) Status(deviceId string) (err error) {
	request := &WsMessage{
		Apikey:   sonoff.ApiKey,
		Action:   "status",
		Deviceid: deviceId,
		Params:   []string{"switch"},
	}
	successCallback := func(device *ConnectedDevice, message *WsMessage) (err error) {
		params := (*message).Params.(map[string]interface{})
		(*device).Status = params["switch"].(string)
		go ws.mqttservice.PublishToStatusTopic(deviceId, (*device).Status)
		return
	}
	err = ws.WriteTo(deviceId, request, successCallback)
	return
}

// DiscardDevice close and discards a ConnectioDevice identified by deviceid
func (ws *WsService) DiscardDevice(deviceId string) {
	if device, ok := ws.devices[deviceId]; ok {
		ws.removeConnectedDevice(device)
	}
}

func (ws *WsService) DeviceById(deviceid string) (device *ConnectedDevice, err error) {
	if value, ok := ws.devices[deviceid]; !ok {
		device = value
	} else {
		err = fmt.Errorf("Unknown device %v", deviceid)
	}
	return
}
