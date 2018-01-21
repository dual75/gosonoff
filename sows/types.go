package sows

import (
	"time"

	"github.com/dual75/gosonoff/sonoff"
)

type WsMessage struct {
	Apikey     string      `json:"apikey"`
	Deviceid   string      `json:"deviceid"`
	Error      int         `json:"error,omitempty"`
	Date       time.Time   `json:"date,omitempty"`
	Action     string      `json:"action,omitempty"`
	Romversion string      `json:"romversion,omitempty"`
	Model      string      `json:"model,omitempty"`
	Params     interface{} `json:"params,omitempty"`
}

func NewWsMessage(deviceid string) *WsMessage {
	return &WsMessage{
		Apikey:   sonoff.ApiKey,
		Deviceid: deviceid,
	}
}
