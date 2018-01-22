package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dual75/gosonoff/sonoff"
)

// configure makes a POST to devices configuration URL to set up HTTP as WS endpoints
func configure(url string, ssid *string, password *string) (err error) {
	if *ssid == "*" {
		err = fmt.Errorf("ssid is mandatory")
		return
	}
	if *password == "*" {
		err = fmt.Errorf("password is mandatory")
		return
	}

	req := map[string]interface{}{
		"version":    4,
		"ssid":       ssid,
		"password":   password,
		"serverName": sonoffConfig.Server.Addr,
		"port":       sonoffConfig.Server.Port,
	}
	body, err := json.Marshal(req)
	var resp *http.Response
	if err == nil {
		sbody := string(body[:])
		fmt.Println("request Body:", sbody)
		resp, err = http.Post(url, sonoff.ContentType, strings.NewReader(sbody))
	}
	if err != nil {
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
	}
	fmt.Println("response Body:", string(body[:]))
	return
}
