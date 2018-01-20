package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/dual75/gosonoff/sonoff"
)

func configure(ssid *string, password *string) (err error) {
	if *ssid == "*" {
		log.Fatal("ssid is mandatory")
	}
	if *password == "*" {
		log.Fatal("password is mandatory")
	}

	req := map[string]interface{}{
		"version":    4,
		"ssid":       ssid,
		"password":   password,
		"serverName": sonoffConfig.Server.Addr,
		"port":       sonoffConfig.Server.Port,
	}
	body, err := json.Marshal(req)
	checkErr(err)

	sbody := string(body[:])
	fmt.Println("request Body:", sbody)
	resp, err := http.Post(sonoff.ConfigurationUrl, sonoff.ContentType, strings.NewReader(sbody))
	checkErr(err)
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	checkErr(err)

	fmt.Println("response Body:", string(body[:]))
	return
}
