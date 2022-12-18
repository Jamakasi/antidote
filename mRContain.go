package main

import (
	"encoding/json"
	"errors"

	"github.com/miekg/dns"
)

type MRContain struct {
	Names []string
	QType string
}

func (p *MRContain) init(data []byte) {
	p.QType = "A"
	if err := json.Unmarshal(data, &p); err != nil {
	}
}
func (p *MRContain) process(data *Data) error {
	if p.isContain(p.Names, data.msg, qTypeToNum(p.QType)) {
		return errors.New("Not found")
	}
	return nil
}

func qTypeToNum(str string) uint16 {
	switch str {
	case "A":
		{
			return dns.TypeA
		}
	case "AAAA":
		{
			return dns.TypeAAAA
		}

	default:
		{
			return dns.TypeA
		}
	}
}

// Проверка по списку адресов с указанием конкретного типа
func (p *MRContain) isContain(targets []string, ans *dns.Msg, dnstype uint16) bool {
	for _, rr := range ans.Answer {
		if rr.Header().Rrtype == dnstype {
			if p.isContainRecord(targets, &rr) {
				return true
			}
		}
	}
	for _, rr := range ans.Extra {
		if rr.Header().Rrtype == dnstype {
			if p.isContainRecord(targets, &rr) {
				return true
			}
		}
	}
	return false
}

// Проверка по списку
func (p *MRContain) isContainRecord(targets []string, rr *dns.RR) bool {
	for i := 1; i == dns.NumField(*rr); i++ {
		ip := dns.Field(*rr, i)
		if contains(targets, ip) {
			return true
		}
	}
	return false
}

// https://play.golang.org/p/Qg_uv_inCek
// contains checks if a string is present in a slice
func (p *MRContain) contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
