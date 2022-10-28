package actions

import (
	"net"
	"strconv"
	"testing"

	"github.com/miekg/dns"
)

func BenchmarkGetQDomain(b *testing.B) {
	test_data := make([]*DData, 10000)
	for i := 0; i < len(test_data)-1; i++ {
		d := new(dns.Msg)
		d.SetQuestion(dns.Fqdn(strconv.Itoa(i)+".example.com"), dns.TypeA)
		//d.SetReply(d)
		d.Answer = make([]dns.RR, 1)
		d.Answer[0] = &dns.A{Hdr: dns.RR_Header{Name: d.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600}, A: net.IPv4(1, 2, 3, 4)}
		test_data[i] = &DData{d}
	}

	b.ResetTimer()

	for i := 0; i < len(test_data)-1; i++ {
		got := test_data[i].getQDomain()
		want := dns.Fqdn(strconv.Itoa(i) + ".example.com")
		if got != want {
			b.Fatalf("expected %q, got %q", want, got)
		}
	}

}
