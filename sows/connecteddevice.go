package sows

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dual75/gosonoff/somqtt"

	"github.com/gorilla/websocket"
)

// You can godoc types

// ReplyCallback is a method to be called upon action request completion
type ReplyCallback func(*ConnectedDevice, *WsMessage) (err error)

// ConnectedDevice is a structure incapsulating device informations and actions
type ConnectedDevice struct {
	Deviceid   string
	Devicetype string
	Romversion string
	Model      string
	Status     string

	mqttservice  somqtt.Publisher
	conn         *websocket.Conn
	replyChannel chan *WsMessage
	writeMux     sync.Mutex
}

// You can godoc methods

func RegisterConnectedDevice(conn *websocket.Conn, mqttservice somqtt.Publisher) (result *ConnectedDevice, err error) {
	msg, err := recvMessage(conn, mqttservice)
	if err == nil {
		if msg.Action == "register" {
			result = &ConnectedDevice{
				Deviceid:     msg.Deviceid,
				Devicetype:   getDeviceType(msg.Deviceid),
				Romversion:   msg.Romversion,
				Model:        msg.Model,
				Status:       "unknown",
				conn:         conn,
				mqttservice:  mqttservice,
				replyChannel: make(chan *WsMessage, 1),
				writeMux:     sync.Mutex{},
			}
			go mqttservice.PublishToStatusTopic(msg.Deviceid, "unknown")
			go mqttservice.SubscribeAll(msg.Deviceid)
		} else {
			err = fmt.Errorf("RegisterConnectedDevice :expected 'register' action, got '%v'", msg.Action)
		}
	}
	if err == nil {
		response := NewWsMessage(msg.Deviceid)
		(*response.Error) = 0
		err = conn.WriteJSON(response)
	}
	return
}

// ServeForever reads from ws in an infinite loop
// new actions are dispatched to handleAction while
func (d *ConnectedDevice) ServeForever() (err error) {
	for {
		msg, err := recvMessage(d.conn, d.mqttservice)
		if err != nil {
			log.Println("ServeForever, error in recvMessage:", err)
			break
		}
		if msg.Action != "" {
			// Current message is an action from device
			response, err := d.handleAction(msg)
			if err != nil {
				log.Println("ServeForever, error in handleAction:", err)
				continue
			}
			err = d.conn.WriteJSON(response)
			if err != nil {
				log.Println("ServeForever, error in WriteJSON:", err)
				break
			}
		} else {
			//
			d.replyChannel <- msg
		}
	}
	return
}

func recvMessage(conn *websocket.Conn, mqttservice somqtt.Publisher) (result *WsMessage, err error) {
	err = conn.ReadJSON(result)
	if err != nil {
		log.Println("recvMessage error:", err)
	}
	return
}

func (d *ConnectedDevice) handleAction(msg *WsMessage) (*WsMessage, error) {
	result := NewWsMessage(msg.Deviceid)
	(*result.Error) = WsReplyOk

	marshaled, err := json.Marshal(msg)
	if err == nil {
		go d.mqttservice.PublishToEventTopic((*msg).Deviceid, marshaled)
		switch msg.Action {
		case "date":
			result.Date = time.Now()
		case "update":
			params := msg.Params.(map[string]interface{})
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
	d.writeMux.Lock()
	defer d.writeMux.Unlock()

	err = d.conn.WriteJSON(data)
	if err == nil {
		reply := <-d.replyChannel
		if callback != nil {
			err = callback(d, reply)
		}
	}
	return
}

func (d *ConnectedDevice) CloseConnection() (err error) {
	err = d.conn.Close()
	return
}
