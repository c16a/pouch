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
	if node.Config.Ws == nil || !node.Config.Ws.Enable {
		return
	}

	var upgrader = websocket.Upgrader{}
	http.Handle("/", handleWsRequest(upgrader, node))

	server := &http.Server{
		Addr:    node.Config.Ws.Addr,
		Handler: http.DefaultServeMux,
	}

	tlsConfig, err := GetTlsConfig(node.Config)
	if err != nil {
		log.Fatal(err)
	} else {
		if tlsConfig != nil {
			server.TLSConfig = tlsConfig
		}
	}

	go func() {
		err := server.ListenAndServeTLS("", "")
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
