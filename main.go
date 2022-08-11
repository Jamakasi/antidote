package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/miekg/dns"
)

// cmd args
var (
	configFile = flag.String("config", "antidote.json", "Configuration file")
	listenAddr = flag.String("listen", "localhost:8053", "Local bind address")
)

func startServer(net string, address string) {
	err := dns.ListenAndServe(address, net, nil)
	if err != nil {
		log.Fatalf("Failed to setup the "+net+" server: %s\n", err.Error())
	}
}

func main() {
	flag.Parse()
	configuration := ReadConfig(*configFile)

	dns.HandleFunc(".", ServerHandler(configuration))

	go startServer("udp", *listenAddr)
	//FIXME ANY query work over tcp
	go startServer("tcp", *listenAddr)

	log.Printf("Started DNS server on: %s\n", *listenAddr)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	recvSig := <-sig
	log.Printf("Signal (%s) received, stopping\n", recvSig.String())
}
