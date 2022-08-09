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

type Record struct {
	Domain  string
	RecType uint16
	Address string
	Ttl     uint32
}

// see dnstype dns.Type
func collectAnswers(ans *dns.Msg, dnstype uint16) []Record {
	/*tmpData := new(Data)
	tmpData.RecType = dnstype
	tmpData.Domain = ans.Question[0].Name*/
	recSlice := make([]Record, 0)

	for _, rr := range ans.Answer {
		if rr.Header().Rrtype == dnstype {
			for i := 1; i <= dns.NumField(rr); i++ {
				recSlice = append(recSlice, Record{Domain: "", Address: dns.Field(rr, i), Ttl: rr.Header().Ttl})
			}
		}
	}
	for _, rr := range ans.Extra {
		if rr.Header().Rrtype == dnstype {
			for i := 1; i <= dns.NumField(rr); i++ {
				recSlice = append(recSlice, Record{Address: dns.Field(rr, i), Ttl: rr.Header().Ttl})
			}
		}
	}

	//tmpData.Records = recSlice
	return recSlice
}

func runActions(ans *dns.Msg, targets []Targets) {
	for _, target := range targets {
		if len(target.A) != 0 {
			data := collectAnswers(ans, dns.TypeA)
			doActions(target.Actions, data)
		}
	}
}

func doActions(actions []Actions, recs []Record) {
	if len(recs) == 0 {
		return
	}

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
			//log.Printf("terminal not implement: %s %s %d", domain, address, ttl)
			if len(action.Actionsr) != 0 {
				doActions(action.Actionsr, recs)
			}
		case "rest-get":
			for _, rec := range recs {
				url := formatTemplate(action.URL, &VarTemplate{Domain: rec.Domain, Address: rec.Address, Ttl: rec.Ttl})
				_, err := http.Get(url)
				if err != nil {
					log.Printf("ERROR rest-get: %s", err)
				}
			}
			if len(action.Actionsr) != 0 {
				doActions(action.Actionsr, recs)
			}
		case "log":
			//Просто вывести в лог содержимое записей
			for _, rec := range recs {
				str := formatTemplate("{{.Domain}} {{.Address}} {{.Ttl}}", &VarTemplate{Domain: rec.Domain, Address: rec.Address, Ttl: rec.Ttl})
				if len(action.STR) != 0 {
					str = formatTemplate(action.STR, &VarTemplate{Domain: rec.Domain, Address: rec.Address, Ttl: rec.Ttl})
				}
				log.Println(str)
			}
			if len(action.Actionsr) != 0 {
				doActions(action.Actionsr, recs)
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
