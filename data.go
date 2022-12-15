package main

import (
	"strings"

	"github.com/miekg/dns"
)

// data
type Data struct {
	msg *dns.Msg
	w   dns.ResponseWriter
}

type ARecord struct {
	Domain  string
	RecType uint16
	Address string
	Ttl     uint32
}

/*return domain name from question with dot at end
 */
func (d *Data) getQDomain() string {
	return dns.Fqdn(d.msg.Question[0].Name)
}

/*
return type from question
See miekg/dns/ztypes.go/TypeToString and miekg/dns/types.go
*/
func (d *Data) getQType() string {
	return dns.TypeToString[d.msg.Question[0].Qtype]
}

/*
return qclass from question
See miekg/dns/msg.go/ClassToString and miekg/dns/types.go
*/
func (d *Data) getQClass() string {
	return dns.ClassToString[d.msg.Question[0].Qclass]
}

// see dnstype dns.Type
// FIXME! ans.Question[0].Name
func (d *Data) collectAnswers(dnstype uint16) []ARecord {
	recSlice := make([]ARecord, 0)

	for _, rr := range d.msg.Answer {
		if rr.Header().Rrtype == dnstype {
			for i := 1; i <= dns.NumField(rr); i++ {
				recSlice = append(recSlice, ARecord{Domain: d.msg.Question[0].Name, Address: dns.Field(rr, i), Ttl: rr.Header().Ttl})
			}
		}
	}
	for _, rr := range d.msg.Extra {
		if rr.Header().Rrtype == dnstype {
			for i := 1; i <= dns.NumField(rr); i++ {
				recSlice = append(recSlice, ARecord{Address: dns.Field(rr, i), Ttl: rr.Header().Ttl})
			}
		}
	}
	return recSlice
}

// see dnstype dns.Type
// FIXME! ans.Question[0].Name
func (d *Data) collectAnswersAll() []ARecord {
	recSlice := make([]ARecord, 0)

	for _, rr := range d.msg.Answer {
		for i := 1; i <= dns.NumField(rr); i++ {
			recSlice = append(recSlice, ARecord{Domain: d.msg.Question[0].Name, Address: dns.Field(rr, i), Ttl: rr.Header().Ttl, RecType: rr.Header().Rrtype})
		}
	}
	for _, rr := range d.msg.Extra {
		for i := 1; i <= dns.NumField(rr); i++ {
			recSlice = append(recSlice, ARecord{Address: dns.Field(rr, i), Ttl: rr.Header().Ttl, RecType: rr.Header().Rrtype})
		}
	}
	return recSlice
}

func (d *Data) appendAddresses(recs []ARecord) string {
	sl := make([]string, 0)
	for _, rec := range recs {
		sl = append(sl, rec.Address)
	}
	return strings.Join(sl, ",")

}
