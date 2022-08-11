package antidote

import (
	"math/rand"
	"time"

	"github.com/miekg/dns"
)

type Resolver struct {
}

func (r *Resolver) Resolve(req *dns.Msg, up *Upstream) (res *dns.Msg, rtt time.Duration, nameserver string, err error) {
	strategy := "random"
	var resp *dns.Msg
	var rttime time.Duration
	var e error
	var ns string
	if len(up.Strategy) != 0 {
		strategy = up.Strategy
	}
	switch strategy {
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

// Взять ns по счетчику, увиличить счетчик
func (r *Resolver) resolveCycle(req *dns.Msg, up *Upstream) (res *dns.Msg, rtt time.Duration, nameserver string, err error) {
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

func (r *Resolver) resolveSingle(req *dns.Msg, nameserver string) (res *dns.Msg, rtt time.Duration, err error) {
	c := new(dns.Client)
	c.Net = "udp"
	return c.Exchange(req, nameserver)
}
