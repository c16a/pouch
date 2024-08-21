package main

import (
	"encoding/json"
	"github.com/c16a/pouch/server/env"
	"github.com/c16a/pouch/server/handlers"
	"github.com/c16a/pouch/server/store"
	"io"
	"log"
	"os"
	"os/signal"
)

func main() {
	configPath := os.Getenv(env.ConfigFilePath)
	if configPath == "" {
		panic("config file path is empty")
	}

	config, err := scanConfigFile[store.NodeConfig](configPath)
	if err != nil {
		panic(err)
	}

	node := store.NewRaftNode(config)
	if err := node.Start(); err != nil {
		log.Fatalf("failed to start node: %s", err.Error())
	}

	go handlers.StartTcpListener(node)
	go handlers.StartWsListener(node)
	go handlers.StartQuicListener(node)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt, os.Kill)
	<-terminate
	log.Println("pouch exiting")
}

func scanConfigFile[T any](path string) (*T, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fileBytes, err := io.ReadAll(configFile)
	if err != nil {
		return nil, err
	}

	var t T
	err = json.Unmarshal(fileBytes, &t)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
