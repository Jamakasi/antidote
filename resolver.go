package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type Resolver struct {
}

func (r *Resolver) Resolve(req *dns.Msg, up *Upstream) (*dns.Msg, time.Duration, string, error) {
	var resp *dns.Msg
	var rttime time.Duration
	var e error
	var ns string

	switch up.Strategy {
	case "random":
		{
			resp, rttime, ns, e = r.resolveRandom(req, up)
		}
	case "cycle":
		{
			resp, rttime, ns, e = r.resolveCycle(req, up)
		}
	case "sequence":
		{
			resp, rttime, ns, e = r.resolveSequence(req, up)
		}
	case "parallel":
		{
			resp, rttime, ns, e = r.resolveParallel(req, up)
		}
	default:
		{
			resp, rttime, ns, e = r.resolveRandom(req, up)
		}
	}
	return resp, rttime, ns, e
}

// Запрос через 1 случайный upstream
func (r *Resolver) resolveRandom(req *dns.Msg, up *Upstream) (*dns.Msg, time.Duration, string, error) {
	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	ns := up.NServers[randGen.Intn(len(up.NServers))]
	resp, rtt, er := r.resolveSingle(req, ns)
	return resp, rtt, ns, er
}

// Запрос через первый, в случае ошибки запрос через следующий пока не кончится список upstream
func (r *Resolver) resolveSequence(req *dns.Msg, up *Upstream) (*dns.Msg, time.Duration, string, error) {
	var resp *dns.Msg
	var rtt time.Duration
	var ns string
	var err error
	var rtt_temp time.Duration
	rtt_temp = 0
	for i := 0; i < len(up.NServers); i++ {
		ns = up.NServers[i]
		var errr error
		resp, rtt_temp, errr = r.resolveSingle(req, ns)
		rtt = rtt_temp + rtt
		if errr == nil {
			err = errr
			break
		}
	}

	return resp, rtt, ns, err
}

// Параллельный запрос через все upstream. Возврат первого быстрейшего без ошибок
func (r *Resolver) resolveParallel(req *dns.Msg, up *Upstream) (*dns.Msg, time.Duration, string, error) {
	type Result struct {
		ans *dns.Msg
		rtt time.Duration
		ns  string
		err error
	}
	wg := new(sync.WaitGroup)
	var results = make(chan Result)
	for _, ns := range up.NServers {
		wg.Add(1)
		go func(req *dns.Msg, ns string, wg *sync.WaitGroup, result chan<- Result) {
			defer wg.Done()
			resp, rtt, er := r.resolveSingle(req, ns)
			result <- Result{ans: resp, rtt: rtt, ns: ns, err: er}
		}(req, ns, wg, results)
	}
	var result Result
	for res := range results {
		if res.err == nil {
			result = res
			go func() {
				<-results
				/*for res := range results {
					//noop, hack to correct close unneeded goroutines
					//log.Printf("unneeded routine end %s , %d ms", res.ns, res.rtt/1e6)
				}*/
			}()
			break
		}
	}
	return result.ans, result.rtt, result.ns, result.err
}

// Взять ns по счетчику, увиличить счетчик
func (r *Resolver) resolveCycle(req *dns.Msg, up *Upstream) (*dns.Msg, time.Duration, string, error) {
	up.CycleMutex.Lock()
	curr := up.CycleCurrent
	if up.CycleCurrent >= len(up.NServers)-1 {
		up.CycleCurrent = 0
	} else {
		up.CycleCurrent = up.CycleCurrent + 1
	}
	up.CycleMutex.Unlock()
	ns := up.NServers[curr]
	resp, rtt, er := r.resolveSingle(req, ns)
	return resp, rtt, ns, er
}

func (r *Resolver) resolveSingle(req *dns.Msg, nameserver string) (*dns.Msg, time.Duration, error) {
	c := new(dns.Client)
	c.Net = "udp"
	return c.Exchange(req, nameserver)
}
