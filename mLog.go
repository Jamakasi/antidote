package main

import (
	"encoding/json"
	"fmt"

	"github.com/miekg/dns"
)

type MLog struct {
	Message string
	Once    bool
}

func (p *MLog) init(data []byte) {
	if err := json.Unmarshal(data, &p); err != nil {
	}
	if len(p.Message) == 0 {
		p.Message = "{{.QDomain}} {{.QType}} {{.QClass}}\n"
	}
}
func (p *MLog) process(data *Data) error {
	recs := data.collectAnswersAll()
	v := &TemplateVars{QDomain: data.getQDomain(), QType: data.getQType(), QClass: data.getQClass(),
		AAllAddress: data.appendAddresses(recs)}
	if p.Once {
		fmt.Println(FormatString(p.Message, v))
	} else {
		for _, rec := range recs {
			v.AAddress = rec.Address
			v.ATtl = rec.Ttl
			v.AType = dns.TypeToString[rec.RecType]
			fmt.Println(FormatString(p.Message, v))
		}
	}
	return nil
}
