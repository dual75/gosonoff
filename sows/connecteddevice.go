package sows

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dual75/gosonoff/somqtt"

	"github.com/gorilla/websocket"
)

// You can godoc types

// ReplyCallback is a method to be called upon action request completion
type ReplyCallback func(*ConnectedDevice, *WsMessage)

// ConnectedDevice is a structure incapsulating device informations and actions
type ConnectedDevice struct {
	Deviceid   string
	Devicetype string
	Romversion string
	Model      string
	Status     string

	mqttservice   *somqtt.MqttService
	conn          *websocket.Conn
	replyCallback ReplyCallback
}

// You can godoc methods

func RegisterConnectedDevice(conn *websocket.Conn, mqttservice *somqtt.MqttService) (result *ConnectedDevice, err error) {
	req, err := recvMessage(conn, mqttservice)
	if err == nil {
		if req.Action == "register" {
			result = &ConnectedDevice{
				Deviceid:    req.Deviceid,
				Devicetype:  getDeviceType(req.Deviceid),
				Romversion:  req.Romversion,
				Model:       req.Model,
				Status:      "unknown",
				conn:        conn,
				mqttservice: mqttservice,
			}
			go mqttservice.PublishToStatusTopic(req.Deviceid, "unknown")
			go mqttservice.SubscribeAll(req.Deviceid)
		} else {
			err = fmt.Errorf("RegisterConnectedDevice :expected 'register' action, got '%v'", req.Action)
		}
	}
	if err == nil {
		response := NewWsMessage(req.Deviceid)
		(*response.Error) = 0
		err = conn.WriteJSON(response)
	}
	return
}

func (d *ConnectedDevice) ServeForever() (err error) {
	for {
		req, err := recvMessage(d.conn, d.mqttservice)
		if err != nil {
			log.Println("ServeForever, error in recvMessage:", err)
			break
		}
		if req.Action != "" {
			// Current message is an action from device
			response, err := d.handleAction(req)
			if err != nil {
				log.Println("ServeForever, error in handleAction:", err)
				continue
			}
			if response != nil {
				err = d.conn.WriteJSON(response)
				if err != nil {
					log.Println("ServeForever, error in WriteJSON:", err)
					break
				}
			}
		} else {
			// if a callback func has been set then execute it
			if d.replyCallback != nil {
				d.replyCallback(d, req)
				// reset replyCallback
				d.replyCallback = nil
			}
		}
	}
	return
}

func recvMessage(conn *websocket.Conn, mqttservice *somqtt.MqttService) (result *WsMessage, err error) {
	err = conn.ReadJSON(result)
	if err == nil {
		go mqttservice.PublishToEventTopic(result.Deviceid, result)
	} else {
		log.Println("recvMessage error:", err)
	}
	return
}

func (d *ConnectedDevice) handleAction(req *WsMessage) (*WsMessage, error) {
	result := NewWsMessage(req.Deviceid)
	(*result.Error) = WsReplyOk

	marshaled, err := json.Marshal(req)
	if err == nil {
		go d.mqttservice.PublishToEventTopic((*req).Deviceid, marshaled)
		switch req.Action {
		case "date":
			result.Date = time.Now()
		case "update":
			params := req.Params.(map[string]interface{})
			status, ok := params["switch"]
			if ok {
				d.Status = status.(string)
				go d.mqttservice.PublishToStatusTopic(d.Deviceid, status.(string))
			}
		case "query":
			result.Params = make(map[string]interface{})
		}
	}
	return result, err
}

func getDeviceType(deviceid string) (result string) {
	prefix := deviceid[:2]
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

func (d *ConnectedDevice) Write(data interface{}, callback ReplyCallback) (err error) {
	err = d.conn.WriteJSON(data)
	if err == nil {
		d.replyCallback = callback
	}
	return
}

func (d *ConnectedDevice) CloseConnection() (err error) {
	err = d.conn.Close()
	return
}
