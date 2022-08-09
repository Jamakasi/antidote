package resolver

import (
	"log"
	"math/rand"
	"time"
	"koozz.ru/antidote"
	"github.com/miekg/dns"
)

func Resolve(w dns.ResponseWriter,req *dns.Msg, up *antidote.Config.Upstream) (r *dns.Msg, rtt time.Duration, hasError bool, nameserver string) {
	strategy := "random"
	var resp *dns.Msg
	var rttime time.Duration
	var err error
	var ns string
	if len(up.Strategy) != 0 {
		strategy = up.Strategy
	}
	switch strategy {
	case "random":
		{
			resp, rttime, err, ns = resolveRandom(req, up)
		}
	case "cycle":
		{
			resp, rttime, err, ns = resolveCycle(req, up)
		}
	default:
		{
			resp, rttime, err, ns = resolveRandom(req, up)
		}
	}
	switch {
	case err != nil:
		log.Printf("ERROR: %s %s\n", ns, err.Error())
		sendFailure(w, req)
		return resp, rtt, true, ns
	case req.Id != resp.Id:
		log.Printf("ERROR: %s Id mismatch: %v != %v\n", ns, req.Id, resp.Id)
		sendFailure(w, req)
		return resp, rtt, true, ns
	}
	return resp, rttime, false, ns
}

// Запрос через 1 случайный upstream
// Ошибка req.Id != resp.Id  log.Printf("ERROR: %s Id mismatch: %v != %v\n", ns, req.Id, resp.Id)
// Ошибка err != nil: log.Printf("ERROR: %s %s\n", ns, err.Error())
func resolveRandom(req *dns.Msg, up *Upstream) (r *dns.Msg, rtt time.Duration, err error, nameserver string) {
	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	ns := up.NServers[randGen.Intn(len(up.NServers))]
	resp, rtt, er := resolveSingle(req, ns)
	return resp, rtt, er, ns
}

//Взять ns по счетчику, увиличить счетчик
func resolveCycle(req *dns.Msg, up *Upstream) (r *dns.Msg, rtt time.Duration, err error, nameserver string) {
	up.CycleMutex.Lock()
	curr := up.CycleCurrent
	if up.CycleCurrent >= len(up.NServers)-1 {
		up.CycleCurrent = 0
	} else {
		up.CycleCurrent = up.CycleCurrent + 1
	}
	up.CycleMutex.Unlock()
	ns := up.NServers[curr]
	resp, rtt, er := resolveSingle(req, ns)
	return resp, rtt, er, ns
}

func resolveSingle(req *dns.Msg, nameserver string) (r *dns.Msg, rtt time.Duration, err error) {
	c := new(dns.Client)
	c.Net = "udp"
	return c.Exchange(req, nameserver)
}
