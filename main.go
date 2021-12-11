package main

import (
	"flag"
	"log"
	"net/http"
	"vephar/web"
)

var (
	httpAddr = flag.String("addr", "127.0.0.1:8080", "HTTP address")
	raftAddr = flag.String("raft", "127.0.0.1:9090", "Raft address")
	confPath = flag.String("conf", "./config.json", "Configuration file path")
)

func parser() {
	flag.Parse()
}

func main() {
	log.Println("Starting smr application")
	parser()

	srv, err := web.NewServer(*confPath, *raftAddr)
	if err != nil {
		log.Fatalf("failed creating server %s: %v", *httpAddr, err)
	}
	http.HandleFunc("/get", srv.GetRequest)
	http.HandleFunc("/set", srv.SetRequest)

	if err := srv.Start(); err != nil {
		log.Fatalf("failed start server %s: %v", *httpAddr, err)
	}

	log.Printf("start listening on [%s]", *httpAddr)
	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}
