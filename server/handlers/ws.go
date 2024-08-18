package handlers

import (
	"github.com/c16a/pouch/sdk/commands"
	"github.com/c16a/pouch/server/store"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
)

func StartWsListener(node *store.RaftNode) {
	var upgrader = websocket.Upgrader{}
	http.Handle("/", handleWsRequest(upgrader, node))

	go func() {
		err := http.ListenAndServe(node.Config.Ws.Addr, http.DefaultServeMux)
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
