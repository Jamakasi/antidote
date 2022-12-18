package main

import (
	"encoding/json"
)

// Example dummy plugin
type MResolve struct {
	NServers []string `json:"ns"`
	Strategy string   `json:"strategy,omitempty"`
}

func (p *MResolve) init(data []byte) {
	if err := json.Unmarshal(data, &p); err != nil {
	}
	if len(p.Strategy) == 0 {
		p.Strategy = "random"
	}
}

func (p *MResolve) process(data *Data) error {
	u := &Upstream{NServers: p.NServers, Strategy: p.Strategy}

	resp, _, _, _ := (new(Resolver)).Resolve(data.msg, u)
	//fmt.Printf("plugin resolve. rtt: %d ns: %s", rtt, resolved_via)
	data.msg = resp
	return nil
}
