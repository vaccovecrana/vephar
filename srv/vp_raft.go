package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/raft"
)

type Server struct {
	peerId    string // host:raftPort:httpPort TODO this may be an issue for IPV6 addresses
	bootPeers map[string]bool
	dataDir   string
	raft      *raft.Raft
	store     *BadgerStore
}

func parsePeer(peer string) (string, string) {
	peerCfg := strings.Split(peer, ":")
	return fmt.Sprintf("%s:%s", peerCfg[0], peerCfg[1]), peerCfg[2]
}

func NewServer(dataDir string, peerId string, bootPeers []string) *Server {
	bootSet := make(map[string]bool)
	bootSet[peerId] = true
	for _, peer := range bootPeers {
		bootSet[peer] = true
	}
	return &Server{peerId: peerId, dataDir: dataDir, bootPeers: bootSet}
}

// This will start the Raft node and will join the cluster after the end.
func (s *Server) Start() error {
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(s.peerId)
	raftAddr, _ := parsePeer(s.peerId)

	addr, err := net.ResolveTCPAddr("tcp", raftAddr)
	if err != nil {
		return err
	}

	trans, err := raft.NewTCPTransport(raftAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	if _, err := os.Stat(s.dataDir); os.IsNotExist(err) {
		if err = os.Mkdir(s.dataDir, 0755); err != nil {
			return err
		}
	}

	badgerdir := fmt.Sprintf("%s/badger", s.dataDir)
	if err := os.MkdirAll(badgerdir, 0755); err != nil {
		return err
	}

	snapshots, err := raft.NewFileSnapshotStore(s.dataDir, 2, os.Stderr)
	if err != nil {
		return err
	}

	bst, err := NewBadgerStore(badgerdir)
	if err != nil {
		return err
	}

	ra, err := raft.NewRaft(raftConfig, bst, bst, bst, snapshots, trans)
	if err != nil {
		return err
	}

	s.raft = ra
	s.store = bst

	var servers []raft.Server
	for peerId := range s.bootPeers {
		log.Info("Registering", "peerId", peerId)
		peerRaft, _ := parsePeer(peerId)
		servers = append(servers, raft.Server{
			ID:      raft.ServerID(peerId),
			Address: raft.ServerAddress(peerRaft),
		})
	}

	// TODO may need to refactor how cluster is initialized.
	// i.e. adding 2 boot peers makes the node stop reacting to heartbeat timeout.
	s.raft.BootstrapCluster(raft.Configuration{Servers: servers})

	return nil
}

func (s *Server) RaftSet(key string, value []byte) error {
	if log.IsDebug() {
		log.Debug("Log Set", "k", key, "v", value)
	}
	command := &VpLogCmd{Op: CMDSET, Key: key, Value: value}
	buff, err := json.Marshal(command)
	if err != nil {
		return err
	}
	future := s.raft.Apply(buff, 10*time.Second)
	if err := future.Error(); err != nil {
		return err
	}
	return nil
}

func (s *Server) RaftDelete(key string) error {
	if log.IsDebug() {
		log.Debug("Log del", "k", key)
	}
	command := &VpLogCmd{Op: CMDDEL, Key: key}
	buff, err := json.Marshal(command)
	if err != nil {
		return err
	}
	future := s.raft.Apply(buff, 10*time.Second)
	if err := future.Error(); err != nil {
		return err
	}
	return nil
}

func (s *Server) RaftJoin(peerId string) error {
	if s.raft.State() != raft.Leader {
		return errors.New("not the leader")
	}
	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}
	peerRaft, _ := parsePeer(peerId)
	f := s.raft.AddVoter(raft.ServerID(peerId), raft.ServerAddress(peerRaft), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	return nil
}

func (s *Server) RaftLeave(peerId string) error {
	if s.raft.State() != raft.Leader {
		return errors.New("not the leader")
	}
	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}
	future := s.raft.RemoveServer(raft.ServerID(peerId), 0, 0)
	if err := future.Error(); err != nil {
		return err
	}
	return nil
}
