package main

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/miekg/dns"
)

type MRest struct {
	HttpMethod         string `json:"method,omitempty"`
	HttpSkipTls        bool   `json:"skiptls,omitempty"`
	HttpBasicAuthLogin string `json:"login,omitempty"`
	HttpBasicAuthPass  string `json:"password,omitempty"`
	Data               string `json:"data,omitempty"`
	URL                string `json:"url,omitempty"`

	Once bool
}

func (p *MRest) init(data []byte) {
	if err := json.Unmarshal(data, &p); err != nil {
	}
}
func (p *MRest) process(data *Data) error {

	return nil
}

func (p *MRest) getQuerry(data *Data) {
	var client *http.Client
	if p.HttpSkipTls {
		transport := &http.Transport{}
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client = &http.Client{Transport: transport}
	} else {
		client = &http.Client{}
	}
	recs := data.collectAnswersAll()
	v := &TemplateVars{QDomain: data.getQDomain(), QType: data.getQType(), QClass: data.getQClass(),
		AAllAddress: url.QueryEscape(data.appendAddresses(recs))}
	if p.Once {
		escurl := FormatString(p.URL, v)
		req, _ := http.NewRequest(p.HttpMethod, escurl, nil)
		if (len(p.HttpBasicAuthLogin) != 0) || (len(p.HttpBasicAuthPass) != 0) {
			req.SetBasicAuth(p.HttpBasicAuthLogin, p.HttpBasicAuthPass)
		}
		responce, _ := client.Do(req)
		defer responce.Body.Close()
	} else {
		for _, rec := range recs {
			v.AAddress = rec.Address
			v.ATtl = rec.Ttl
			v.AType = dns.TypeToString[rec.RecType]
			escurl := FormatString(p.URL, v)
			req, _ := http.NewRequest(p.HttpMethod, escurl, nil)
			if (len(p.HttpBasicAuthLogin) != 0) || (len(p.HttpBasicAuthPass) != 0) {
				req.SetBasicAuth(p.HttpBasicAuthLogin, p.HttpBasicAuthPass)
			}
			responce, _ := client.Do(req)
			defer responce.Body.Close()
		}

	}
}
