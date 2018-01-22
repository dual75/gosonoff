// sows package hold all the type and functions involved in websocket communication between server and devices

package sows

import (
	"time"

	"github.com/dual75/gosonoff/sonoff"
)

// You can godoc types

// WsMessage represents a JSON message for server <-> device communication
type WsMessage struct {
	Apikey     string      `json:"apikey"`
	Deviceid   string      `json:"deviceid"`
	Error      *int        `json:"error,omitempty"`
	Date       time.Time   `json:"date,omitempty"`
	Action     string      `json:"action,omitempty"`
	Romversion string      `json:"romversion,omitempty"`
	Model      string      `json:"model,omitempty"`
	Params     interface{} `json:"params,omitempty"`
}

// You can godoc functions

// NewWsMessage creates a new WsMessage with given deviceid and default api key
func NewWsMessage(deviceid string) *WsMessage {
	return &WsMessage{
		Apikey:   sonoff.ApiKey,
		Deviceid: deviceid,
	}
}
