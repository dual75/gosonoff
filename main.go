// main is the entry point for gosonoff

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/dual75/gosonoff/sonoff"
	"github.com/go-yaml/yaml"
)

// You can godoc functions

// checkErr check for last error and exit with log.Fatal
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// parseArgs configure parser and parse command line arguments
func parseArgs() (*string, *string, *string, *string, *string, *string) {
	config := flag.String("config", sonoff.ConfigFile, "configuration file name")
	command := flag.String("command", sonoff.CommandServe, "command to execute [serve, configure]")
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

	err = yaml.Unmarshal(data, &sonoff.Config)
	checkErr(err)

	switch *command {
	case sonoff.CommandServe:
		err = serve(certfile, keyfile)
	case sonoff.CommandConfigure:
		err = configure(sonoff.ConfigurationUrl, ssid, password)
	default:
		err = fmt.Errorf("Unknown command: %v", command)
	}
	checkErr(err)
}
