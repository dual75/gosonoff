package sohttp

type SonoffHttp struct {
	Addr string
	Port int
}

type WebSocketConfig struct {
	Error  int    `json:"error"`
	Reason string `json:"reason"`
	IP     string `json:"IP"`
	Port   int    `json:"port"`
}
