package models

import (
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type WsManager struct {
	sync.RWMutex
	Clients map[*websocket.Conn]bool
}

func (wm *WsManager) AddClient(c *websocket.Conn) {
	wm.Lock()
	defer wm.Unlock()
	wm.Clients[c] = true
}

func (wm *WsManager) RemoveClient(c *websocket.Conn) {
	wm.Lock()
	defer wm.Unlock()
	delete(wm.Clients, c)
}
