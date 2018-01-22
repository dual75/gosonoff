// main is the entry point for gosonoff

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/dual75/gosonoff/sohttp"
	"github.com/dual75/gosonoff/somqtt"
	"github.com/dual75/gosonoff/sonoff"
	"github.com/go-yaml/yaml"
)

// SonoffConfig holds configuration from yaml file
type SonoffConfig struct {
	Server sohttp.SonoffHttp
	Mqtt   somqtt.SonoffMqtt
}

var sonoffConfig SonoffConfig = SonoffConfig{}

// checkErr check for last error and exit with log.Fatal
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// parseArgs configure parser and parse command line arguments
func parseArgs() (*string, *string, *string, *string, *string, *string) {
	config := flag.String("config", sonoff.ConfigFile, "configuration file name")
	command := flag.String("command", sonoff.CommandDefault, "command to execute [serve, configure]")
	certfile := flag.String("cert", sonoff.CertFile, "server certificate file")
	keyfile := flag.String("key", sonoff.KeyFile, "server certificate key file")
	ssid := flag.String("ssid", "*", "network ssid")
	password := flag.String("password", "*", "network password")
	flag.Parse()
	return config, command, certfile, keyfile, ssid, password
}

func main() {
	config, command, certfile, keyfile, ssid, password := parseArgs()
	data, err := ioutil.ReadFile(*config)
	checkErr(err)

	err = yaml.Unmarshal(data, &sonoffConfig)
	checkErr(err)

	vCommand := *command
	switch vCommand {
	case sonoff.CommandDefault:
		err = serve(certfile, keyfile)
	case "configure":
		err = configure(sonoff.ConfigurationUrl, ssid, password)
	default:
		err = fmt.Errorf("Unknown command: %v", command)
	}
	checkErr(err)
}
