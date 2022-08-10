package antidote

import (
	"bytes"
	"crypto/tls"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/miekg/dns"
)

type Job struct {
}

type Record struct {
	Domain  string
	RecType uint16
	Address string
	Ttl     uint32
}

// see dnstype dns.Type
// FIXME! ans.Question[0].Name
func (j *Job) collectAnswers(ans *dns.Msg, dnstype uint16) []Record {
	recSlice := make([]Record, 0)

	for _, rr := range ans.Answer {
		if rr.Header().Rrtype == dnstype {
			for i := 1; i <= dns.NumField(rr); i++ {
				recSlice = append(recSlice, Record{Domain: ans.Question[0].Name, Address: dns.Field(rr, i), Ttl: rr.Header().Ttl})
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

func (j *Job) RunActions(ans *dns.Msg, server *Server) {
	for _, target := range server.Targets {
		if len(target.A) != 0 {
			data := j.collectAnswers(ans, dns.TypeA)
			j.doActions(target.Actions, data)
		}
	}
}

func (j *Job) doActions(actions []Action, recs []Record) {
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
			if len(action.Actions) != 0 {
				j.doActions(action.Actions, recs)
			}
		case "rest-get":
			for _, rec := range recs {
				url := formatTemplate(action.URL, &VarTemplate{Domain: rec.Domain, Address: rec.Address, Ttl: rec.Ttl})
				_, err := http.Get(url)
				if err != nil {
					log.Printf("ERROR rest-get: %s", err)
				}
			}
			if len(action.Actions) != 0 {
				j.doActions(action.Actions, recs)
			}
		case "log":
			//Просто вывести в лог содержимое записей
			jobLog(action, recs)
			if len(action.Actions) != 0 {
				j.doActions(action.Actions, recs)
			}
		case "rest":
			jobRest(action, recs)
			if len(action.Actions) != 0 {
				j.doActions(action.Actions, recs)
			}
		}
	}
}

func jobLog(act Action, recs []Record) {
	for _, rec := range recs {
		str := formatTemplate("{{.Domain}} {{.Address}} {{.Ttl}}", &VarTemplate{Domain: rec.Domain, Address: rec.Address, Ttl: rec.Ttl})
		if len(act.STR) != 0 {
			str = formatTemplate(act.STR, &VarTemplate{Domain: rec.Domain, Address: rec.Address, Ttl: rec.Ttl})
		}
		log.Println(str)
	}
}

func jobRest(act Action, recs []Record) {
	if len(act.URL) == 0 {
		log.Println("ERROR ACTION : rest \"url\" not set!")
		return
	}
	if len(act.HttpMethod) == 0 {
		log.Println("ERROR ACTION : rest \"method\" not set!")
		return
	}
	if act.HttpMethod == "GET" {
		for _, rec := range recs {
			url := formatTemplate(act.URL, &VarTemplate{Domain: rec.Domain, Address: rec.Address, Ttl: rec.Ttl})
			_, err := http.Get(url)
			if err != nil {
				log.Printf("ERROR rest: %s", err.Error())
				return
			}
		}
		return
	}
	if len(act.Data) == 0 {
		log.Println("ERROR ACTION : rest \"data\" not set!")
		return
	}
	var client *http.Client

	log.Printf("act.HttpSkipTls %t", act.HttpSkipTls)
	if act.HttpSkipTls {
		transport := &http.Transport{}
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client = &http.Client{Transport: transport}
	} else {
		client = &http.Client{}
	}

	for _, rec := range recs {
		data := []byte(formatTemplate(act.Data, &VarTemplate{Domain: rec.Domain, Address: rec.Address, Ttl: rec.Ttl}))
		log.Printf("req : %s", data)
		req, err := http.NewRequest(act.HttpMethod, act.URL, bytes.NewBuffer(data))
		if err != nil {
			log.Printf("ERROR rest in prepare request: %s", err.Error())
			return
		}
		req.Header.Set("Content-Type", "application/json")
		if (len(act.HttpBasicAuthLogin) != 0) || (len(act.HttpBasicAuthPass) != 0) {
			req.SetBasicAuth(act.HttpBasicAuthLogin, act.HttpBasicAuthPass)
		}
		_, errr := client.Do(req)
		if errr != nil {
			log.Printf("ERROR rest: %s", errr.Error())
			return
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
