package antidote

import (
	"encoding/json"
	"io/ioutil"
	"log"
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
	Domain  string
	Address string
	Ttl     uint32
}

type Config struct {
	Server Server `json:"server"`
}
type Actions struct {
	//может быть у любого
	Type     string    `json:"type"`
	Actionsr []Actions `json:"actions,omitempty"`
	//terminal
	Cmd string `json:"cmd,omitempty"`
	//rest-get
	URL string `json:"url,omitempty"`
	//log
	STR string `json:"str,omitempty"`
}
type Targets struct {
	A       []string  `json:"A,omitempty"`
	AAAA    []string  `json:"AAAA,omitempty"`
	Actions []Actions `json:"actions"`
}
type Server struct {
	DNSbad  []string `json:"dns_bad"`
	DNSgood []string `json:"dns_good"`
	//	IgnoreDomains []string `json:"ignore_domains,omitempty"`
	Targets []Targets `json:"targets"`
}

func ReadConfig(filename string) Config {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Can't open config file: %s", err.Error())
	}

	var jsonConfig Config
	if err := json.Unmarshal(file, &jsonConfig); err != nil {
		log.Fatalf("Cannot parse the configuration: %s", err)
	}

	// Safety checks
	if len(jsonConfig.Server.DNSgood) == 0 {
		log.Fatal("Configuration contains no 'dns_good' section")
	}
	if len(jsonConfig.Server.DNSbad) == 0 {
		log.Fatal("Configuration contains no 'dns_bad' section")
	}
	return jsonConfig
}
