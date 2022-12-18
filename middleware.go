package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"strings"
)

// Abstract plugin
type MiddlewareBase struct {
	Type      string            `json:"type"`
	pInstance MiddlewareIface   //targed plugin instance
	RawData   json.RawMessage   `json:"config,omitempty"`
	OnOk      []*MiddlewareBase `json:"onok,omitempty"`
	OnFail    []*MiddlewareBase `json:"onfail,omitempty"`
}
type MiddlewareIface interface {
	init(data []byte)
	process(data *Data) error
}

func (p *MiddlewareBase) do(data *Data) {
	err := p.pInstance.process(data)
	if err == nil {
		for _, next := range p.OnOk {
			next.do(data)
		}
	} else {
		for _, next := range p.OnFail {
			next.do(data)
		}
	}

}

// Рекусрвиная инициализация плагинов.
// Передача им куска json с их конфигами
func (p *MiddlewareBase) initPluginsRecursive() {
	switch p.Type {
	case "log":
		{
			plug := MLog{}
			plug.init(p.RawData)
			p.pInstance = &plug
		}
	case "rest":
		{
			plug := MRest{}
			plug.init(p.RawData)
			p.pInstance = &plug
		}
	case "resolve":
		{
			plug := MResolve{}
			plug.init(p.RawData)
			p.pInstance = &plug
		}
	case "sendResponce":
		{
			plug := MSendResponce{}
			plug.init(p.RawData)
			p.pInstance = &plug
		}
	case "rcontain":
		{
			plug := MRContain{}
			plug.init(p.RawData)
			p.pInstance = &plug
		}
	case "dummy":
		{
			plug := MDummy{}
			plug.init(p.RawData)
			p.pInstance = &plug
		}
	default:
		{
			fmt.Printf("Unknown plugin type: %s\n", p.Type)
		}
	}
	for _, next := range p.OnOk {
		next.initPluginsRecursive()
	}
	for _, next := range p.OnFail {
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
