package main

import (
	"github.com/c16a/pouch/server/env"
	"github.com/c16a/pouch/server/handlers"
	"github.com/c16a/pouch/server/store"
	"log"
	"os"
	"os/signal"
)

func main() {
	peerAddr := os.Getenv(env.PeerAddr)
	enableSingle := false
	if peerAddr == "" {
		enableSingle = true
	}

	node := store.NewRaftNode()
	if err := node.Start(enableSingle); err != nil {
		log.Fatalf("failed to start node: %s", err.Error())
	}

	go handlers.StartTcpListener(node)
	go handlers.StartWsListener(node)
	//go handlers.StartQuicListener(node)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt, os.Kill)
	<-terminate
	log.Println("pouch exiting")
}
