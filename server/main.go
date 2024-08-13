package main

import (
	"flag"
	"fmt"
	"github.com/c16a/pouch/server/env"
	"github.com/c16a/pouch/server/handlers"
	"github.com/c16a/pouch/server/peering"
	"github.com/c16a/pouch/server/store"
	"github.com/google/uuid"
	"log"
	"os"
	"os/signal"
)

// Command line parameters
func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <raft-data-path> \n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	raftPath := os.Getenv(env.RaftDir)
	if raftPath == "" {
		log.Fatalf("Environment variable %s is not set\n", env.RaftDir)
	}
	if err := os.MkdirAll(raftPath, 0700); err != nil {
		log.Fatalf("failed to create path for Raft storage: %s", err.Error())
	}

	raftAddr := os.Getenv(env.RaftAddr)
	if raftAddr == "" {
		log.Fatalf("Environment variable %s is not set\n", env.RaftAddr)
	}

	peerAddr := os.Getenv(env.PeerAddr)
	enableSingle := false
	if peerAddr == "" {
		enableSingle = true
	}

	nodeId := os.Getenv(env.NodeId)
	if nodeId == "" {
		id, err := uuid.NewV7()
		if err != nil {
			log.Fatalf("failed to generate node id: %s", err.Error())
		}
		nodeId = id.String()
	}

	node := store.New()
	node.RaftDir = raftPath
	node.RaftBind = raftAddr
	if err := node.Open(enableSingle, nodeId); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}
	go peering.InitPeer(nodeId, peerAddr, node)
	go handlers.StartTcpListener(node)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt, os.Kill)
	<-terminate
	log.Println("pouch exiting")
}
