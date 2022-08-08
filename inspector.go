package antidote

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/miekg/dns"
)

func isPoisoned(ans *dns.Msg, targets []Targets) bool {
	for _, target := range targets {
		//A records
		if len(target.A) != 0 {
			for _, rr := range ans.Answer {
				if rr.Header().Rrtype == dns.TypeA {
					if isMatch(target.A, &rr) {
						return true
					}
				}
			}
			for _, rr := range ans.Extra {
				if rr.Header().Rrtype == dns.TypeA {
					if isMatch(target.A, &rr) {
						return true
					}
				}
			}
		}
		//AAAA records
		if len(target.AAAA) != 0 {
			for _, rr := range ans.Answer {
				if rr.Header().Rrtype == dns.TypeAAAA {
					if isMatch(target.AAAA, &rr) {
						return true
					}
				}
			}
			for _, rr := range ans.Extra {
				if rr.Header().Rrtype == dns.TypeAAAA {
					if isMatch(target.AAAA, &rr) {
						return true
					}
				}
			}
		}
	}

	return false
}

func isMatch(targets []string, rr *dns.RR) bool {
	for i := 1; i == dns.NumField(*rr); i++ {
		ip := dns.Field(*rr, i)
		if contains(targets, ip) {
			return true
		}
	}
	return false
}

func runActions(ans *dns.Msg, targets []Targets) {
	for _, target := range targets {
		if len(target.A) != 0 {
			for _, rr := range ans.Answer {
				if rr.Header().Rrtype == dns.TypeA {
					for i := 1; i <= dns.NumField(rr); i++ {
						doActions(target.Actions, ans.Question[0].Name, dns.Field(rr, i), rr.Header().Ttl)
					}

				}
			}
		}
	}
}

func doActions(actions []Actions, domain string, address string, ttl uint32) {
	for _, action := range actions {
		switch action.Type {
		case "terminal":
			/*cmd := exec.Command(*command, *fqdn, rec)
			var out bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr
			err := cmd.Run()
			if err != nil {
				text := "runCommand: " + fmt.Sprint(err) + ": " + stderr.String()
				log.Error(text)
			}*/
			log.Printf("terminal not implement: %s %s %d", domain, address, ttl)
			if len(action.Actionsr) != 0 {
				doActions(action.Actionsr, domain, address, ttl)
			}
		case "rest-get":
			url := formatTemplate(action.URL, &VarTemplate{Domain: domain, Address: address, Ttl: ttl})
			_, err := http.Get(url)
			if err != nil {
				log.Printf("ERROR get: %s", err)
			}
			if len(action.Actionsr) != 0 {
				doActions(action.Actionsr, domain, address, ttl)
			}
		case "log":
			str := formatTemplate("{{.Domain}} {{.Address}} {{.Ttl}}", &VarTemplate{Domain: domain, Address: address, Ttl: ttl})
			if len(action.STR) != 0 {
				str = formatTemplate(action.STR, &VarTemplate{Domain: domain, Address: address, Ttl: ttl})
			}
			log.Println(str)
			if len(action.Actionsr) != 0 {
				doActions(action.Actionsr, domain, address, ttl)
			}
		}
	}
}

func formatTemplate(s string, v interface{}) string {
	strbuild := new(strings.Builder)
	tmpl, err := template.New("test").Parse(s)
	if err != nil {
		log.Printf("Error formatTemplate parse %s %s : %s", s, v, err.Error())
	}
	err = tmpl.Execute(strbuild, v)
	if err != nil {
		log.Printf("Error formatTemplate parse %s %s : %s", s, v, err.Error())
	}
	return strbuild.String()
	/*t, b := new(template.Template), new(strings.Builder)
	template.Must(t.Parse(s)).Execute(b, v)
	return b.String()*/
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
