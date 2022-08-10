package antidote

import (
	"log"
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
		//resolve
		resp_bad, rtt_bad, ns_bad, err_bad := (new(Resolver)).Resolve(req, &config.Server.UpstreamBad)
		//stop if errors
		if checkErrors(w, req, resp_bad, ns_bad, err_bad) {
			return
		}
		if !isPoisoned(resp_bad, config.Server.Targets) {
			sendResponse(w, resp_bad, rtt_bad, ns_bad, err_bad)
		} else {
			resp_good, rtt_good, ns_good, err_good := (new(Resolver)).Resolve(req, &config.Server.UpstreamGood)
			//stop if errors
			if checkErrors(w, req, resp_good, ns_good, err_good) {
				return
			}
			sendResponse(w, resp_good, rtt_good, ns_good, err_good)
			(new(Job)).RunActions(resp_good, &config.Server)
		}

	} // end of handler
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
