package antidote

import (
	"log"
	"math/rand"
	"time"

	"github.com/miekg/dns"
)

// Handler is the handler function that will serve DNS requests.
type Handler func(dns.ResponseWriter, *dns.Msg)

// ServerHandler Returns an anonymous function configured to resolve DNS
// queries with a specific set of remote servers.
func ServerHandler(config Config) Handler {
	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))

	// This is the actual handler
	return func(w dns.ResponseWriter, req *dns.Msg) {

		/*for _, q := range req.Question {
			log.Printf("Incoming request #%v: %s %s %v - using %s\n",
				req.Id,
				dns.ClassToString[q.Qclass],
				dns.TypeToString[q.Qtype],
				q.Name, nameserver)
		}*/

		c := new(dns.Client)
		c.Net = "udp"
		//Запрос через плохой ns сервер
		nameserver_bad := config.Server.UpstreamBad[randGen.Intn(len(config.Server.UpstreamBad))]
		resp_bad, rtt_bad, has_err := resolve(w, req, c, nameserver_bad)
		if !has_err {
			if !isPoisoned(resp_bad, config.Server.Targets) {
				//Нет совпадений. Просто отвечаем.
				log.Printf("Request #%v: %d ms, server: %s(%s), size: %d bytes\n", resp_bad.Id, rtt_bad/1e6, nameserver_bad, c.Net, resp_bad.Len())
				if err := w.WriteMsg(resp_bad); err != nil {
					log.Printf("ERROR: %s write failed: %s", nameserver_bad, err)
				}

			} else {
				//Совпадение. Делаем запрос через доверенные сервера, если все хорошо то засылаем ответ
				nameserver_good := config.Server.UpstreamGood[randGen.Intn(len(config.Server.UpstreamGood))]
				resp_good, rtt_good, has_err_good := resolve(w, req, c, nameserver_good)
				if !has_err_good {
					//Ошибок нет. Отправляем ответ и выполняем actions
					log.Printf("Request #%v: %d ms, server: %s(%s), size: %d bytes\n", resp_good.Id, rtt_good/1e6, nameserver_good, c.Net, resp_good.Len())
					if err := w.WriteMsg(resp_good); err != nil {
						log.Printf("ERROR: %s write failed: %s", nameserver_good, err)
					}
					runActions(resp_good, config.Server.Targets)
				}
			}
		}

	} // end of handler
}

func resolveNew(w dns.ResponseWriter, req *dns.Msg, configs Server) (r *dns.Msg, rtt time.Duration, hasError bool) {
	strategy := "random"
	resp := new(dns.Msg)
	rttime := new(time.Duration)
	err := nil
	if len(configs.UpstreamStrategy) != 0 {
		strategy = configs.UpstreamStrategy
	}
	switch strategy {
	case "random":
		{
			resolveRandom
		}
	default:
		{

		}
	}
	switch {
	case err != nil:
		log.Printf("ERROR: %s %s\n", ns, err.Error())
		sendFailure(w, req)
		return resp, rtt, true
	case req.Id != resp.Id:
		log.Printf("ERROR: %s Id mismatch: %v != %v\n", ns, req.Id, resp.Id)
		sendFailure(w, req)
		return resp, rtt, true
	}
	return resp, rtt, false
}

// Запрос через 1 случайный upstream
// Ошибка req.Id != resp.Id  log.Printf("ERROR: %s Id mismatch: %v != %v\n", ns, req.Id, resp.Id)
// Ошибка err != nil: log.Printf("ERROR: %s %s\n", ns, err.Error())
func resolveRandom(req *dns.Msg, nameservers []string) (r *dns.Msg, rtt time.Duration, err error) {
	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	c := new(dns.Client)
	c.Net = "udp"
	//Запрос через плохой ns сервер
	ns := nameservers[randGen.Intn(len(nameservers))]
	return c.Exchange(req, ns)
}

func resolve(w dns.ResponseWriter, req *dns.Msg, client *dns.Client, ns string) (r *dns.Msg, rtt time.Duration, hasError bool) {
	resp, rtt, err := client.Exchange(req, ns)
	switch {
	case err != nil:
		log.Printf("ERROR: %s %s\n", ns, err.Error())
		sendFailure(w, req)
		return resp, rtt, true
	case req.Id != resp.Id:
		log.Printf("ERROR: %s Id mismatch: %v != %v\n", ns, req.Id, resp.Id)
		sendFailure(w, req)
		return resp, rtt, true
	}
	return resp, rtt, false
}

func sendFailure(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetRcode(r, dns.RcodeServerFailure)
	if err := w.WriteMsg(msg); err != nil {
		log.Printf("ERROR: write failed in sendFailure: %s", err)
	}
}
