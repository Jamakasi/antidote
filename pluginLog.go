package main

import (
	"encoding/json"
	"fmt"

	"github.com/miekg/dns"
)

type PluginLog struct {
	Message string
}

func (p *PluginLog) init(data []byte) PluginIface {
	var plog PluginLog
	if err := json.Unmarshal(data, &plog); err != nil {
		plog = PluginLog{}
	}
	if len(plog.Message) == 0 {
		plog.Message = "{{.Domain}} {{.Address}} {{.Ttl}}"
	}
	return p
}
func (p *PluginLog) process(w dns.ResponseWriter, req *dns.Msg) {
	fmt.Println(p.Message)
}
