package main

import "encoding/json"

// Example dummy plugin
type MDummy struct {
}

func (p *MDummy) init(data []byte) {
	if err := json.Unmarshal(data, &p); err != nil {
	}
}
func (p *MDummy) process(data *Data) error {
	return nil
}
