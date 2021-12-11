package web

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/raft"
)

type peer struct {
	Raft string `json:raft`
	Http string `json:http`
}

type clusterConfig struct {
	Peers []peer `json:peers`
	Dir   string `json:dir`
}

type Server struct {
	bind   string
	config clusterConfig
	raft   *raft.Raft
}

// This will start the Raft node and will join the cluster after the end.
func (s *Server) Start() error {
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(s.bind)

	addr, err := net.ResolveTCPAddr("tcp", s.bind)
	if err != nil {
		return err
	}

	trans, err := raft.NewTCPTransport(s.bind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	if _, err := os.Stat(s.config.Dir); os.IsNotExist(err) {
		if err = os.Mkdir(s.config.Dir, 0755); err != nil {
			return err
		}
	}

	completeDir := s.config.Dir + "/" + s.bind
	if _, err := os.Stat(completeDir); os.IsNotExist(err) {
		if err = os.Mkdir(completeDir, 0755); err != nil {
			return err
		}
	}

	snapshots, err := raft.NewFileSnapshotStore(completeDir, 2, os.Stderr)
	if err != nil {
		return err
	}

	mem := raft.NewInmemStore()

	dmap := NewDatabase()
	ra, err := raft.NewRaft(raftConfig, dmap, mem, mem, snapshots, trans)
	if err != nil {
		return err
	}

	s.raft = ra
	var servers []raft.Server
	for _, peer := range s.config.Peers {
		servers = append(servers, raft.Server{
			ID:      raft.ServerID(peer.Raft),
			Address: raft.ServerAddress(peer.Raft),
		})
	}
	s.raft.BootstrapCluster(raft.Configuration{Servers: servers})

	return nil
}

// @deprecated
func (s *Server) Join(peer string) error {
	raftConfig := s.raft.GetConfiguration()
	if err := raftConfig.Error(); err != nil {
		return err
	}

	for _, server := range raftConfig.Configuration().Servers {
		if server.ID == raft.ServerID(peer) || server.Address == raft.ServerAddress(peer) {
			fmt.Printf("node %s is already a member\n", peer)
			return nil
		}

		future := s.raft.RemoveServer(server.ID, 0, 0)
		if future.Error() != nil {
			return future.Error()
		}
	}

	future := s.raft.AddVoter(raft.ServerID(peer), raft.ServerAddress(peer), 0, 0)
	if future.Error() != nil {
		return future.Error()
	}
	fmt.Printf("node %s joined cluster", peer)
	return nil
}

func NewServer(confPath string, raftBind string) (*Server, error) {
	buf, err := ioutil.ReadFile(confPath)
	if err != nil {
		return nil, err
	}

	var config clusterConfig
	if err = json.Unmarshal(buf, &config); err != nil {
		return nil, err
	}

	return &Server{config: config, bind: raftBind}, nil
}

func (s *Server) GetRequest(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	key := request.Form.Get("key")

	if s.raft.State() != raft.Leader {
		s.forwardToLeader(writer, request)
		return
	}

	command := &Command{
		Op:  "get",
		Key: key,
	}

	buff, err := json.Marshal(command)
	if err != nil {
		writer.WriteHeader(500)
		fmt.Fprintf(writer, "error = %v", err)
		return
	}

	future := s.raft.Apply(buff, 10*time.Second)

	if err := future.Error(); err != nil {
		writer.WriteHeader(500)
		fmt.Fprintf(writer, "error = %v", err)
		return
	}

	writer.WriteHeader(200)
	fmt.Fprintf(writer, "[key = %s, value = %v]", key, future.Response())
}

func (s *Server) SetRequest(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	key := request.Form.Get("key")
	value := request.Form.Get("value")

	if s.raft.State() != raft.Leader {
		s.forwardToLeader(writer, request)
		return
	}

	command := &Command{
		Op:    "set",
		Key:   key,
		Value: value,
	}

	buff, err := json.Marshal(command)
	if err != nil {
		writer.WriteHeader(500)
		fmt.Fprintf(writer, "error = %v", err)
		return
	}

	future := s.raft.Apply(buff, 10*time.Second)
	if err := future.Error(); err != nil {
		writer.WriteHeader(500)
		fmt.Fprintf(writer, "apply error = %s", err)
		return
	}

	writer.WriteHeader(200)
	fmt.Fprintf(writer, "[key = %s, value = %s]", key, value)
}

func (s *Server) forwardToLeader(writer http.ResponseWriter, request *http.Request) {
	var url string
	for _, peer := range s.config.Peers {
		if raft.ServerAddress(peer.Raft) == s.raft.Leader() {
			url = peer.Http
		}
	}

	if len(url) == 0 {
		writer.WriteHeader(500)
		fmt.Fprint(writer, "leader not found")
		return
	}

	url = url + request.RequestURI
	fmt.Printf("forwarding request from %s to %s\n", s.bind, url)

	res, err := http.Get(url)
	if err != nil {
		writer.WriteHeader(500)
		fmt.Fprintf(writer, "error forward: %v", err)
		return
	}

	defer res.Body.Close()
	io.Copy(writer, res.Body)
}
