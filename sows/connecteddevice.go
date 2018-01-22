package sows

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dual75/gosonoff/somqtt"

	"github.com/gorilla/websocket"
)

type ReplyCallback func(*ConnectedDevice, *WsMessage)

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

func RegisterConnectedDevice(conn *websocket.Conn, mqttservice *somqtt.MqttService) (result *ConnectedDevice, err error) {
	req, err := recvMessage(conn, mqttservice)
	var vreq WsMessage
	if err == nil {
		vreq = *req
		if vreq.Action == "register" {
			result = &ConnectedDevice{
				Deviceid:    vreq.Deviceid,
				Devicetype:  getDeviceType(vreq.Deviceid),
				Romversion:  vreq.Romversion,
				Model:       vreq.Model,
				Status:      "unknown",
				conn:        conn,
				mqttservice: mqttservice,
			}
			go mqttservice.PublishToStatusTopic(vreq.Deviceid, "unknown")
			go mqttservice.SubscribeAll(vreq.Deviceid)
		} else {
			err = fmt.Errorf("expected 'register' action, got '%v'", vreq.Action)
		}
	}
	if err == nil {
		response := NewWsMessage(vreq.Deviceid)
		(*response.Error) = 0
		err = conn.WriteJSON(response)
	}
	return
}

func (d *ConnectedDevice) ServeForever() (err error) {
	for {
		req, err := recvMessage(d.conn, d.mqttservice)
		if err != nil {
			log.Println("error in recvMessage:", err)
			break
		}
		if req.Action != "" {
			// Current message is an action from device
			response, err := d.handleAction(req)
			if err != nil {
				log.Println("error in handleAction:", err)
				continue
			}
			if response != nil {
				err = d.conn.WriteJSON(response)
				if err != nil {
					log.Println("serveForever, error in WriteJSON:", err)
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
	if err != nil {
		log.Printf("recvMessage error: %v", err)
	} else {
		go mqttservice.PublishToEventTopic((*result).Deviceid, result)
	}
	return
}

func (d *ConnectedDevice) handleAction(req *WsMessage) (*WsMessage, error) {
	result := NewWsMessage(req.Deviceid)
	(*result.Error) = 0

	marshaled, err := json.Marshal(req)
	if err == nil {
		go d.mqttservice.PublishToEventTopic((*req).Deviceid, marshaled)
		switch req.Action {
		case "date":
			(*result).Date = time.Now()
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
