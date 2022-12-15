package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"strings"
)

// Abstract plugin
type PluginBase struct {
	Type         string          `json:"type"`
	pInstance    PluginIface     //targed plugin instance
	RawData      json.RawMessage `json:"config,omitempty"`
	Plugins      []*PluginBase   `json:"plugins,omitempty"`
	ErrorPlugins []*PluginBase   `json:"errorPlugins,omitempty"`
}
type PluginIface interface {
	init(data []byte)
	process(data *Data) error
}

func (p *PluginBase) do(data *Data) {
	err := p.pInstance.process(data)
	if err != nil {
		for _, next := range p.Plugins {
			next.do(data)
		}
	} else {
		for _, next := range p.ErrorPlugins {
			next.do(data)
		}
	}

}

// Рекусрвиная инициализация плагинов.
// Передача им куска json с их конфигами
func (p *PluginBase) initPluginsRecursive() {
	switch p.Type {
	case "log":
		{
			plug := PluginLog{}
			plug.init(p.RawData)
			p.pInstance = &plug
		}
	case "rest":
		{
			plug := PluginRest{}
			plug.init(p.RawData)
			p.pInstance = &plug
		}
	case "dummy":
		{
			plug := PluginDummy{}
			plug.init(p.RawData)
			p.pInstance = &plug
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
type TemplateVars struct {
	QDomain string
	QType   string
	QClass  string

	AAddress    string
	ATtl        uint32
	AType       string
	AAllAddress string
}

// Генерация строки по заданному патерну s
// FIXME валится при ошибках в s!!!
func FormatString(s string, v *TemplateVars) string {
	strbuild := new(strings.Builder)
	tmpl, err := template.New("test").Parse(s)
	if err != nil {
		log.Printf("Error FormatString prepare %s : %s", s, err.Error())
	}
	err = tmpl.Execute(strbuild, v)
	if err != nil {
		log.Printf("Error FormatString execute %v : %s", v, err.Error())
	}
	return strbuild.String()
}
