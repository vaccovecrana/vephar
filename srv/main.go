package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/go-hclog"
)

const (
	version = "v0.5.0"
)

var (
	peerId  = flag.String("peerId", "", "host:raftPort:httpPort")
	dataDir = flag.String("data", "", "Data storage directory")
	join    = flag.String("join", "", "Comma-separated list of host:raftPort:httpPort cluster nodes")
	log     = hclog.New(&hclog.LoggerOptions{Name: "vephar"})
)

func main() {
	if _, trace := os.LookupEnv("VPR_TRACE"); trace {
		log.SetLevel(hclog.Trace)
	} else if _, debug := os.LookupEnv("VPR_DEBUG"); debug {
		log.SetLevel(hclog.Debug)
	}

	log.Info("", "version", version)
	flag.Parse()

	if len(os.Args) < 3 {
		log.Error("error: wrong number of arguments")
		flag.Usage()
	} else {
		srv := NewServer(*dataDir, *peerId, strings.Split(*join, ","))
		if err := srv.Start(); err != nil {
			log.Error("failed to start server", "peerId", *peerId, "error", err)
		}

		hdl := NewWebHandler(srv)
		http.HandleFunc(RKvList, hdl.KeysRequest)
		http.HandleFunc(RKvGet, hdl.GetRequest)
		http.HandleFunc(RKvSet, hdl.SetRequest)
		http.HandleFunc(RKvDel, hdl.DeleteRequest)
		http.HandleFunc(RRfJoin, hdl.RaftJoinRequest)
		http.HandleFunc(RRfLeave, hdl.RaftLeaveRequest)
		http.HandleFunc(RRfStat, hdl.RaftStatusRequest)
		http.HandleFunc(RUi, ResourceHandler)
		http.HandleFunc(RIndexJs, ResourceHandler)
		http.HandleFunc(RIndexCss, ResourceHandler)
		http.HandleFunc(RFavIcon, ResourceHandler)
		http.HandleFunc("/", UiHandler)

		peerRaft, httpPort := parsePeer(*peerId)
		peerHttp := fmt.Sprintf("%s:%s", strings.Split(peerRaft, ":")[0], httpPort)
		log.Info("Peer started", "peerId", *peerId)
		log.Error("", "status", http.ListenAndServe(peerHttp, nil))
	}
}
