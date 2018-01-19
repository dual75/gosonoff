package sohttp

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/dual75/gosonoff/sonoff"
)

type WebSocketConfig struct {
	Error  int
	Reason string
	IP     string
	Port   int
}

// ServeHTTP Single handle func
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip, port, err := parseAddr(r)
	if err != nil {
		log.Fatal(err)
	}
	wsConfig := WebSocketConfig{0, "ok", ip, port}
	w.Header().Set("Content-Type", sonoff.ContentType)
	json.NewEncoder(w).Encode(wsConfig)
}

func parseAddr(r *http.Request) (ip string, port int, err error) {
	chunks := strings.SplitN(r.Host, ":", 2)
	if len(chunks) == 2 {
		port, err = strconv.Atoi(chunks[1])
	} else {
		log.Fatal("can't determine port in addr:" + r.Host)
	}
	if err == nil {
		ip = chunks[0]
	}
	return
}
