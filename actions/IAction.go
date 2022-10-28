package actions

import (
	"encoding/json"

	"github.com/miekg/dns"
)

type IAction interface {
	run()
	parseParams(raw json.RawMessage) IAction
}

type Action struct {
	Type      string          `json:"type"`
	params    json.RawMessage `json:"params"`
	onSuccess []Action        `json:"onSuccess,omitempty"`
	onError   []Action        `json:"onError,omitempty"`
	onFail    []Action        `json:"onFail,omitempty"`
}

func (a *Action) run() {
}
func (a *Action) parseParams(raw json.RawMessage) IAction {
}

type DData struct {
	msg *dns.Msg
}

/*return domain name from question with dot at end
 */
func (data *DData) getQDomain() string {
	return dns.Fqdn(data.msg.Question[0].Name)
}

/*
return type from question
See miekg/dns/ztypes.go/TypeToString and miekg/dns/types.go
*/
func (data *DData) getQType() string {
	return dns.TypeToString[data.msg.Question[0].Qtype]
}

/*
return qclass from question
See miekg/dns/msg.go/ClassToString and miekg/dns/types.go
*/
func (data *DData) getQClass() string {
	return dns.ClassToString[data.msg.Question[0].Qclass]
}

type Record struct {
	RecType uint16
	Address string
	Ttl     uint32
}

func getRecords(msg *dns.Msg, rtype uint16) []Record {
	recSlice := make([]Record, 0)

	for _, rr := range msg.Answer {
		if rr.Header().Rrtype == rtype {
			for i := 1; i <= dns.NumField(rr); i++ {
				recSlice = append(recSlice, Record{Address: dns.Field(rr, i), Ttl: rr.Header().Ttl})
			}
		}
	}
	for _, rr := range msg.Extra {
		if rr.Header().Rrtype == rtype {
			for i := 1; i <= dns.NumField(rr); i++ {
				recSlice = append(recSlice, Record{Address: dns.Field(rr, i), Ttl: rr.Header().Ttl})
			}
		}
	}
	return recSlice
}

/*
return slice of Record contain A addresses from answer and extra sections
*/
func (data *DData) getAddressesA() []Record {
	return getRecords(data.msg, dns.TypeA)
}

/*
return slice of Record contain AAAA addresses from answer and extra sections
*/
func (data *DData) getAddressesAAAA() []Record {
	return getRecords(data.msg, dns.TypeAAAA)
}

func (data *DData) parseJson() {

}
