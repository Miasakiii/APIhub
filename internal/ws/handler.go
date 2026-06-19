package ws

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for local development
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handler returns a Gin handler that upgrades HTTP to WebSocket.
func Handler(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("[ws] upgrade error: %v", err)
			return
		}

		client := &Client{
			hub:  hub,
			conn: conn,
			send: make(chan *Message, 64),
		}

		hub.register <- client

		// Start pumps in separate goroutines
		go client.writePump()
		go client.readPump()
	}
}
