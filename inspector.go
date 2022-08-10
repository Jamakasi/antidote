package antidote

import (
	"github.com/miekg/dns"
)

func isPoisoned(ans *dns.Msg, targets []Targets) bool {
	for _, target := range targets {
		//A records
		if len(target.A) != 0 {
			if isPoison(target.A, ans, dns.TypeA) {
				return true
			}
		}
		//AAAA records
		if len(target.AAAA) != 0 {
			if isPoison(target.AAAA, ans, dns.TypeAAAA) {
				return true
			}
		}
	}

	return false
}

// Проверка по списку адресов с указанием конкретного типа
func isPoison(targets []string, ans *dns.Msg, dnstype uint16) bool {
	for _, rr := range ans.Answer {
		if rr.Header().Rrtype == dnstype {
			if isMatchRecord(targets, &rr) {
				return true
			}
		}
	}
	for _, rr := range ans.Extra {
		if rr.Header().Rrtype == dnstype {
			if isMatchRecord(targets, &rr) {
				return true
			}
		}
	}
	return false
}

// Проверка по списку
func isMatchRecord(targets []string, rr *dns.RR) bool {
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
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
