package main

import (
	"encoding/json"
	"github.com/c16a/pouch/sdk/logging"
	"github.com/c16a/pouch/server/env"
	"github.com/c16a/pouch/server/handlers"
	"github.com/c16a/pouch/server/store"
	"go.uber.org/zap"
	"io"
	"log"
	"os"
	"os/signal"
)

func main() {
	logger, err := logging.SetupLogger()
	if err != nil {
		log.Fatal(err)
	}

	configPath := os.Getenv(env.ConfigFilePath)
	if configPath == "" {
		logger.Fatal("config file path is empty")
	}

	config, err := scanConfigFile[store.NodeConfig](configPath)
	if err != nil {
		logger.Fatal("error reading config file", zap.Error(err))
	}

	node := store.NewRaftNode(config, logger)
	if err := node.Start(); err != nil {
		logger.Fatal("error starting node", zap.Error(err))
	}

	go handlers.StartTcpListener(node)
	go handlers.StartWsListener(node)
	go handlers.StartQuicListener(node)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt, os.Kill)
	<-terminate
	logger.Info("pouch exiting")
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
