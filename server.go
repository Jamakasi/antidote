package main

import (
	"log"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// Handler is the handler function that will serve DNS requests.
type Handler func(dns.ResponseWriter, *dns.Msg)

// ServerHandler Returns an anonymous function configured to resolve DNS
func ServerHandler(config *Config) Handler {
	// This is the actual handler
	return func(w dns.ResponseWriter, req *dns.Msg) {
		/*for _, q := range req.Question {
			log.Printf("Incoming request #%v: %s %s %v - using %s\n",
				req.Id,
				dns.ClassToString[q.Qclass],
				dns.TypeToString[q.Qtype],
				q.Name, nameserver)
		}*/
		if config.Server.Parallel {
			parallel(w, req, &config.Server)
		} else {
			sequence(w, req, &config.Server)
		}
		log.Println("end handler")
	} // end of handler
}

func parallel(w dns.ResponseWriter, req *dns.Msg, server *Server) {
	type Result struct {
		ans  *dns.Msg
		rtt  time.Duration
		ns   string
		err  error
		good bool
	}
	wg := new(sync.WaitGroup)
	var results = make(chan *Result)

	wg.Add(1)
	go func(reqq *dns.Msg, up *Upstream, wg *sync.WaitGroup, result chan<- *Result) {
		defer wg.Done()
		resp, rtt, ns, er := (new(Resolver)).Resolve(reqq, up)
		result <- &Result{ans: resp, rtt: rtt, ns: ns, err: er, good: false}
	}(req, &server.UpstreamBad, wg, results)
	wg.Add(1)
	go func(reqq *dns.Msg, up *Upstream, wg *sync.WaitGroup, result chan<- *Result) {
		defer wg.Done()
		resp, rtt, ns, er := (new(Resolver)).Resolve(reqq, up)
		result <- &Result{ans: resp, rtt: rtt, ns: ns, err: er, good: true}
	}(req, &server.UpstreamGood, wg, results)

	var good_result, bad_result *Result
	for result := range results {
		if !result.good {
			bad_result = result
		} else {
			good_result = result
		}
	}
	if checkErrors(w, req, bad_result.ans, bad_result.ns, bad_result.err) {
		return
	}
	if !isPoisoned(bad_result.ans, server.Targets) {
		sendResponse(w, bad_result.ans, bad_result.rtt, bad_result.ns, bad_result.err)
	} else {
		if checkErrors(w, req, good_result.ans, good_result.ns, good_result.err) {
			return
		}
		sendResponse(w, good_result.ans, good_result.rtt, good_result.ns, good_result.err)
		(new(Job)).RunActions(good_result.ans, server)
	}
}
func sequence(w dns.ResponseWriter, req *dns.Msg, server *Server) {
	//resolve
	resp_bad, rtt_bad, ns_bad, err_bad := (new(Resolver)).Resolve(req, &server.UpstreamBad)
	//stop if errors
	if checkErrors(w, req, resp_bad, ns_bad, err_bad) {
		return
	}
	if !isPoisoned(resp_bad, server.Targets) {
		sendResponse(w, resp_bad, rtt_bad, ns_bad, err_bad)
	} else {
		resp_good, rtt_good, ns_good, err_good := (new(Resolver)).Resolve(req, &server.UpstreamGood)
		//stop if errors
		if checkErrors(w, req, resp_good, ns_good, err_good) {
			return
		}
		sendResponse(w, resp_good, rtt_good, ns_good, err_good)
		(new(Job)).RunActions(resp_good, server)
	}
}

func checkErrors(w dns.ResponseWriter, req *dns.Msg, resp *dns.Msg, ns string, err error) bool {
	if err != nil {
		log.Printf("ERROR: %s %s\n", ns, err.Error())
		sendFailure(w, req)
		return true
	}
	if req.Id != resp.Id {
		log.Printf("ERROR: %s Id mismatch: %v != %v\n", ns, req.Id, resp.Id)
		sendFailure(w, req)
		return true
	}
	return false
}

func sendResponse(w dns.ResponseWriter, resp *dns.Msg, rtt time.Duration, ns string, err error) {
	log.Printf("Request #%v: %d ms, server: %s, size: %d bytes\n", resp.Id, rtt/1e6, ns, resp.Len())
	if err := w.WriteMsg(resp); err != nil {
		log.Printf("ERROR: %s write failed: %s", ns, err)
	}
}

func sendFailure(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetRcode(r, dns.RcodeServerFailure)
	if err := w.WriteMsg(msg); err != nil {
		log.Printf("ERROR: write failed in sendFailure: %s", err)
	}
}
