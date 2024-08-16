package handlers

import (
	"github.com/c16a/pouch/sdk/commands"
	"github.com/c16a/pouch/server/env"
	"github.com/c16a/pouch/server/store"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"strings"
)

func StartWsListener(node *store.RaftNode) {

	wsAddr := os.Getenv(env.WsAddr)
	if wsAddr == "" {
		log.Fatalf("Environment variable %s not set", env.WsAddr)
	}

	var upgrader = websocket.Upgrader{}
	http.Handle("/", handleWsRequest(upgrader, node))

	go func() {
		err := http.ListenAndServe(wsAddr, http.DefaultServeMux)
		if err != nil {
			log.Fatal(err)
		}
	}()
}

func handleWsRequest(upgrader websocket.Upgrader, node *store.RaftNode) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				continue
			}
			if mt == websocket.CloseMessage {
				break
			}

			line := strings.TrimSpace(string(message))
			cmd, err := commands.ParseStringIntoCommand(line)
			if err != nil {
				continue
			}

			response := node.ApplyCmd(cmd)

			c.WriteMessage(websocket.TextMessage, []byte(response+"\n"))
		}
	})
}
