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

// /// testing
// filter user
type ClientInfo struct {
	Conn       *websocket.Conn
	Department string
}

type ProjectClientInfo struct {
	Conn      *websocket.Conn
	UserID    string
	Branch    string
	ProjectId string
}

type WsManager2 struct {
	sync.RWMutex
	Clients map[*websocket.Conn]ClientInfo
}

type WsProjectManager struct {
	sync.RWMutex
	ProjectInfo map[*websocket.Conn]ProjectClientInfo
}

func (wm *WsProjectManager) AddProjectClient(c *websocket.Conn, userid string, branch string, projectid string) {
	wm.Lock()
	defer wm.Unlock()
	if wm.ProjectInfo == nil {
		wm.ProjectInfo = make(map[*websocket.Conn]ProjectClientInfo)
	}

	wm.ProjectInfo[c] = ProjectClientInfo{
		Conn:      c,
		UserID:    userid,
		Branch:    branch,
		ProjectId: projectid,
	}
}

func (wm *WsProjectManager) RemoveProjectClient(c *websocket.Conn) {
	wm.Lock()
	defer wm.Unlock()
	delete(wm.ProjectInfo, c)
}

func (wm *WsManager2) AddClient2(c *websocket.Conn, department string) {
	wm.Lock()
	defer wm.Unlock()
	if wm.Clients == nil {
		wm.Clients = make(map[*websocket.Conn]ClientInfo) // Defensive init
	}
	wm.Clients[c] = ClientInfo{
		Conn:       c,
		Department: department,
	}
}

func (wm *WsManager2) RemoveClient2(c *websocket.Conn) {
	wm.Lock()
	defer wm.Unlock()
	delete(wm.Clients, c)
}
