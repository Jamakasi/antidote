package main

import (
	"encoding/json"
	"log"
	"os"
)

/*
 {
	"server": {
		"dns_bad" : ["8.8.4.4:53" , "8.8.8.8:53"],
		"dns_good" : ["1.1.1.1:53" , "1.0.0.1:53"],
		"ignore_domains": ["lawfilter.ertelecom.ru"],
		"targets": [
			{
				"A": ["1.2.3.4", "5.6.7.8"],
				"actions" : [
								{
									"type" : "log",
									"format" : ""
									"actions" : [
										{
											"type" : "rest-get",
											"url" : "http:example.com/api?domain={{.Domain}}&address={{.Address}}&ttl={{.Ttl}}"
										}
									]
								}
							]
			},
			{
				"AAAA": ["abc::1", "def::1"],
				"actions" : [
								{
									"type" : "terminal",
									"cmd" : "myscript.sh {{.Domain}} {{.Address}} {{.Ttl}}"
								},
								{
									"type" : "rest-get",
									"url" : "http:example.com/api?domain={{.Domain}}&address={{.Address}}&ttl={{.Ttl}}"
								}
							]
			}
		]
	}
 }
*/

type Config struct {
	Server Server `json:"server"`
}
type RawJsonAction struct {
	Type      string            `json:"type"`
	params    json.RawMessage   `json:"params"`
	onSuccess []json.RawMessage `json:"onSuccess,omitempty"`
	onError   []json.RawMessage `json:"onError,omitempty"`
	onFail    []json.RawMessage `json:"onFail,omitempty"`
}
type Server struct {
	Ingress RawJsonAction `json:"Ingress"`
}

func ReadConfig(filename string) *Config {
	file, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Can't open config file: %s", err.Error())
	}

	var jsonConfig Config
	if err := json.Unmarshal(file, &jsonConfig); err != nil {
		log.Fatalf("Cannot parse the configuration: %s", err)
	}

	// Safety checks
	/*if len(jsonConfig.Server.UpstreamGood) == 0 {
		log.Fatal("Configuration contains no 'dns_good' section")
	}
	if len(jsonConfig.Server.UpstreamBad) == 0 {
		log.Fatal("Configuration contains no 'dns_bad' section")
	}*/
	return &jsonConfig
}
