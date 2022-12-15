package main

import "encoding/json"

// Example dummy plugin
type PluginResolve struct {
	NServers []string `json:"ns"`
	Strategy string   `json:"strategy,omitempty"`
}

func (p *PluginResolve) init(data []byte) {
	if err := json.Unmarshal(data, &p); err != nil {
	}
	if len(p.Strategy) == 0 {
		p.Strategy = "random"
	}
}

func (p *PluginResolve) process(data *Data) error {

	//resp_bad, rtt_bad, ns_bad, err_bad := (new(Resolver)).Resolve(req, &server.UpstreamBad)
	return nil
}
