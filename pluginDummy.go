package main

import "encoding/json"

// Example dummy plugin
type PluginDummy struct {
}

func (p *PluginDummy) init(data []byte) {
	if err := json.Unmarshal(data, &p); err != nil {
	}
}
func (p *PluginDummy) process(data *Data) error {
	return nil
}
