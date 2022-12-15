package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/miekg/dns"
)

// data
type Data struct {
	message *dns.Msg
	w       dns.ResponseWriter
}

func (d *Data) getQ()

// Abstract plugin
type PluginBase struct {
	Type         string          `json:"type"`
	pInstance    PluginIface     //targed plugin instance
	RawData      json.RawMessage `json:"config,omitempty"`
	Plugins      []*PluginBase   `json:"plugins,omitempty"`
	ErrorPlugins []*PluginBase   `json:"errorPlugins,omitempty"`
}
type PluginIface interface {
	init(data []byte) PluginIface
	process(w dns.ResponseWriter, req *dns.Msg)
}

func (p *PluginBase) do(w dns.ResponseWriter, req *dns.Msg) {
	p.pInstance.process(w, req)
	for _, next := range p.Plugins {
		next.do(w, req)
	}
	for _, next := range p.ErrorPlugins {
		next.do(w, req)
	}
}

// Рекусрвиная инициализация плагинов.
// Передача им куска json с их конфигами
func (p *PluginBase) initPluginsRecursive() {
	switch p.Type {
	case "log":
		{
			/*var plog PluginLog
			if err := json.Unmarshal(p.RawData, &plog); err != nil {
				//log.Fatalf("Cannot parse the configuration: %s", err)
				plog = PluginLog{}
			}*/
			plog := PluginLog{}
			plog.init(p.RawData)
			p.pInstance = &plog
		}
	default:
		{
			fmt.Printf("Unknown plugin type: %s\n", p.Type)
		}
	}
	for _, next := range p.Plugins {
		next.initPluginsRecursive()
	}
	for _, next := range p.ErrorPlugins {
		next.initPluginsRecursive()
	}
}

// Темплейт
type TmpTemplate struct {
	Domain     string
	Address    string
	Ttl        uint32
	AllAddress string
	PrevError  string
}

// Генерация строки по заданному патерну s
func FormatString(s string) string {
	v := &TmpTemplate{}
	strbuild := new(strings.Builder)
	tmpl, err := template.New("test").Parse(s)
	if err != nil {
		log.Printf("Error formatTemplate parse %s %v : %s", s, v, err.Error())
	}
	err = tmpl.Execute(strbuild, v)
	if err != nil {
		log.Printf("Error formatTemplate parse %s %v : %s", s, v, err.Error())
	}
	return strbuild.String()
}
