package main

import (
	"bytes"
	"crypto/tls"
	"html/template"
	"log"
	"net/http"
	"net/url"
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
	Error   string
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
		if len(target.AAAA) != 0 {
			data := j.collectAnswers(ans, dns.TypeAAAA)
			j.doActions(target.Actions, data)
		}
		if len(target.HTTP_REDIRECT_TEST) != 0 {
			dataA := j.collectAnswers(ans, dns.TypeA)
			dataAAAA := j.collectAnswers(ans, dns.TypeAAAA)
			dataA = append(dataA, dataAAAA...)
			j.doActions(target.Actions, dataA)
		}
	}
}

func (j *Job) doActions(actions []Action, recs []Record) {
	if len(recs) == 0 {
		return
	}

	for _, action := range actions {
		//log.Printf("Iteration action : %+v", action)
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

/*
Вывод в лог.
*/
func jobLog(act Action, recs []Record) {
	allAddreses := genAllAddress(recs)
	for _, rec := range recs {
		str := formatTemplate("{{.Domain}} {{.Address}} {{.Ttl}}", rec, allAddreses)
		if len(act.STR) != 0 {
			str = formatTemplate(act.STR, rec, allAddreses)
		}
		log.Println(str)
		if act.Once {
			return
		}
	}
}

/*
Выполнить rest запрос.
Может вернуть массив зафейленных запросов.
Вернет nil если ошибок нет, в противном случае массив записей с ошибками для отработки цепочки action при ошибке
*/
func jobRest(act Action, recs []Record) []Record {
	allAddreses := genAllAddress(recs)
	errRecs := make([]Record, 0)
	if len(act.URL) == 0 {
		log.Println("ERROR ACTION : rest \"url\" not set!")
		return
	}
	if len(act.HttpMethod) == 0 {
		log.Println("ERROR ACTION : rest \"method\" not set!")
		return
	}
	var client *http.Client
	if act.HttpSkipTls {
		transport := &http.Transport{}
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client = &http.Client{Transport: transport}
	} else {
		client = &http.Client{}
	}
	for _, rec := range recs {
		var req *http.Request
		var err error
		if act.HttpMethod == "GET" {
			escurl := formatTemplate(act.URL, rec, url.QueryEscape(allAddreses))
			req, err = http.NewRequest(act.HttpMethod, escurl, nil)
		} else {
			data := []byte(formatTemplate(act.Data, rec, allAddreses))
			req, err = http.NewRequest(act.HttpMethod, act.URL, bytes.NewBuffer(data))
			req.Header.Set("Content-Type", "application/json")
		}

		if err != nil {
			log.Printf("ERROR rest in prepare request: %s", err.Error())
			return
		}

		if (len(act.HttpBasicAuthLogin) != 0) || (len(act.HttpBasicAuthPass) != 0) {
			req.SetBasicAuth(act.HttpBasicAuthLogin, act.HttpBasicAuthPass)
		}
		_, errr := client.Do(req)
		//FIX ME! при ошибке выпадет вверх не отработав остальные записи
		if errr != nil {
			log.Printf("ERROR rest: %s", errr.Error())
			return
		}
		if act.Once {
			return
		}
	}
}

func genAllAddress(recs []Record) string {
	sl := make([]string, 0)
	for _, rec := range recs {
		sl = append(sl, rec.Address)
	}
	return strings.Join(sl, ",")
}

func formatTemplate(s string, rec Record, allAddreses string) string {
	v := &VarTemplate{Domain: rec.Domain, Address: rec.Address, Ttl: rec.Ttl, AllAddress: allAddreses, PrevError: rec.Error}
	strbuild := new(strings.Builder)
	tmpl, err := template.New("test").Parse(s)
	if err != nil {
		log.Printf("Error formatTemplate parse %s %v : %s", s, v, err.Error())
	}
	err = tmpl.Execute(strbuild, v)
	if err != nil {
		log.Printf("Error formatTemplate parse %s %v : %s", s, v, err.Error())
	}
	return strbuild.String()
}
