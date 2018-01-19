package sohttp

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dual75/gosonoff/sonoff"
)

type WebSocketConfig struct {
	Error  int    `json:"error"`
	Reason string `json:"reason"`
	IP     string `json:"IP"`
	Port   int    `json:"port"`
}

type HTTPServer struct {
	Ip     string
	Port   int
	Wsport int
}

// ServeHTTP Single handle func
func (server HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wsConfig := WebSocketConfig{0, "ok", server.Ip, server.Wsport}
	log.Printf("request: URI = %v", r.RequestURI)
	bytes, err := ioutil.ReadAll(r.Body)
	if err == nil {
		log.Printf("request: body = %v", string(bytes))
	} else {
		log.Println(err)
	}
	w.Header().Set("Content-Type", sonoff.ContentType)
	json.NewEncoder(w).Encode(wsConfig)
}
