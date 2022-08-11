package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
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

type VarTemplate struct {
	Domain     string
	Address    string
	Ttl        uint32
	AllAddress string
}

type Config struct {
	Server Server `json:"server"`
}
type Action struct {
	//может быть у любого
	Type    string   `json:"type"`
	Actions []Action `json:"actions,omitempty"`
	Once    bool     `json:"once,omitempty"`
	//terminal
	Cmd string `json:"cmd,omitempty"`
	//log
	STR string `json:"str,omitempty"`
	//rest
	HttpMethod         string `json:"method,omitempty"`
	HttpSkipTls        bool   `json:"skiptls,omitempty"`
	HttpBasicAuthLogin string `json:"login,omitempty"`
	HttpBasicAuthPass  string `json:"password,omitempty"`
	Data               string `json:"data,omitempty"`
	URL                string `json:"url,omitempty"`
}
type Targets struct {
	A       []string `json:"A,omitempty"`
	AAAA    []string `json:"AAAA,omitempty"`
	Actions []Action `json:"actions"`
}
type Upstream struct {
	NServers     []string   `json:"ns"`
	Strategy     string     `json:"strategy,omitempty"`
	CycleMutex   sync.Mutex // mytex для циклической стратегии
	CycleCurrent int        // указатель на текущий сервер для циклической стратегии
}
type Server struct {
	UpstreamBad  Upstream  `json:"upstream_bad"`
	UpstreamGood Upstream  `json:"upstream_good"`
	Parallel     bool      `json:"parallel"`
	Targets      []Targets `json:"targets"`
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
