package actions

import (
	"encoding/json"
	"log"
)

type StdLog struct {
	template string `json:"template"`
}

func (a *StdLog) run() {
}
func (a *StdLog) parseParams(raw json.RawMessage) IAction {
	var action StdLog
	if err := json.Unmarshal(raw, &action); err != nil {
		log.Fatalf("Cannot parse the configuration: %s", err)
	}
	return &action
}
