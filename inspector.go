package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

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
		//Хак для тех кто не блочит днс, а устраивает редирект по http на уровне L7 .Ростелеком
		if len(target.HTTP_REDIRECT_TEST) != 0 {
			if isPoisonHTTP_Redir_Test(target.HTTP_REDIRECT_TEST, ans) {
				return true
			}
		}
	}

	return false
}

func isPoisonHTTP_Redir_Test(targets []string, ans *dns.Msg) bool {
	var (
		dnsResolverIP        = "8.8.8.8:53" // Google DNS resolver.
		dnsResolverProto     = "udp"        // Protocol to use for the DNS resolver
		dnsResolverTimeoutMs = 5000         // Timeout (ms) for the DNS resolver (optional)
	)
	dialer := &net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Duration(dnsResolverTimeoutMs) * time.Millisecond,
				}
				return d.DialContext(ctx, dnsResolverProto, dnsResolverIP)
			},
		},
	}

	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, addr)
	}

	http.DefaultTransport.(*http.Transport).DialContext = dialContext
	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://"+ans.Question[0].Name, nil)
	if err != nil {
		log.Printf("test NewRequest: %s", err.Error())
	}
	//client := new(http.Client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return errors.New("Redirect")
	}

	response, err := client.Do(req)
	//log.Printf("response: status :%s, %v", response.Status, response)
	if err != nil {
		log.Printf("err: status :%s", err.Error())
	}
	if err == nil {
		if response.StatusCode == http.StatusFound { //status code 302
			fmt.Println(response.Location())
			loc, _ := response.Location()
			for _, v := range targets {
				if strings.Contains(loc.String(), v) {
					return true
				}
			}
		}
	} else {
		for _, v := range targets {
			if strings.Contains(err.Error(), v) {
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
