package main

import (
	"encoding/json"
	"log"
)

// Example dummy plugin
type MSendResponce struct {
}

func (p *MSendResponce) init(data []byte) {
	if err := json.Unmarshal(data, &p); err != nil {
	}
}
func (p *MSendResponce) process(data *Data) error {
	if data.w != nil {
		if err := data.w.WriteMsg(data.msg); err != nil {
			log.Printf("ERROR: write failed: %s", err)
		}
		data.w.Close()
		data.w = nil
	}
	return nil
}
