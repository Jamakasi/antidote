package main

import (
	"bytes"
	"crypto/tls"
	"html/template"
	"io"
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
			okR, erR := jobRest(action, recs)
			if okR != nil && len(action.Actions) != 0 {
				j.doActions(action.Actions, okR)
			}
			if erR != nil && len(action.ErrorActions) != 0 {
				j.doActions(action.ErrorActions, erR)
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
Возвращает успешные и неуспешные Records. В случае отсутсвия nil
Заполняет поле Error ошибкой
*/
func jobRest(act Action, recs []Record) ([]Record, []Record) {
	allAddreses := genAllAddress(recs)
	errRecs := make([]Record, 0)
	okRecs := make([]Record, 0)
	if len(act.URL) == 0 {
		return nil, func(r []Record, s string) []Record {
			result := make([]Record, 0)
			for _, rec := range recs {
				r := rec //copy
				r.Error = s
				result = append(result, r)
			}
			return result
		}(recs, "rest \"url\" not set!")
	}
	if len(act.HttpMethod) == 0 {
		return nil, func(r []Record, s string) []Record {
			result := make([]Record, 0)
			for _, rec := range recs {
				r := rec //copy
				r.Error = s
				result = append(result, r)
			}
			return result
		}(recs, "rest \"method\" not set!")
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
			//log.Printf("ERROR rest in prepare request: %s", err.Error())
			return nil, func(r []Record, s string) []Record {
				result := make([]Record, 0)
				for _, rec := range recs {
					r := rec //copy
					r.Error = s
					result = append(result, r)
				}
				return result
			}(recs, err.Error())
		}

		if (len(act.HttpBasicAuthLogin) != 0) || (len(act.HttpBasicAuthPass) != 0) {
			req.SetBasicAuth(act.HttpBasicAuthLogin, act.HttpBasicAuthPass)
		}
		responce, errr := client.Do(req)
		if errr != nil {
			//log.Printf("ERROR rest: %s", errr.Error())
			r := rec //copy
			r.Error = errr.Error()
			errRecs = append(errRecs, r)
		} else {
			defer responce.Body.Close()
			if responce.StatusCode == http.StatusOK {
				okRecs = append(okRecs, rec)
			} else {
				r := rec //copy
				bodyBytes, err := io.ReadAll(responce.Body)
				if err != nil {
					log.Fatal(err)
				}
				if err != nil {
					r.Error = responce.Status + "read body error"
				} else {
					r.Error = responce.Status + string(bodyBytes)
				}
				errRecs = append(errRecs, r)
			}

		}
		if act.Once {
			return okRecs, errRecs
		}
	}
	return okRecs, errRecs
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
