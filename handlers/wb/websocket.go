package handlers

import (
	"net/http"

	"github.com/pierceperado/smpc/initializers/ws/hub"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func ServeWebSocket(h *hub.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	channel := r.URL.Query().Get("channel")
	if channel == "" {
		channel = "default"
	}

	client := &hub.Client{
		Conn:    conn,
		Send:    make(chan []byte),
		Channel: channel,
	}

	h.Register <- client

	go client.WritePump()
	go client.ReadPump(h)
}
