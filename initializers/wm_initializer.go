package initializers

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/pierceperado/smpc/models"
)

var WM *models.WsManager
var WM2 *models.WsManager2
var WM3 *models.WsProjectManager

var WMQuotation *models.WsManager
var WMJobOrder *models.WsManager

func InitWmQuotation() {
	WMQuotation = &models.WsManager{
		Clients: make(map[*websocket.Conn]bool),
	}
}

func InitWmJobOrder() {
	WMJobOrder = &models.WsManager{
		Clients: make(map[*websocket.Conn]bool),
	}
}

func InitWm() {
	WM = &models.WsManager{
		Clients: make(map[*websocket.Conn]bool),
	}
}

// redbox
func InitWm2() {
	WM2 = &models.WsManager2{
		Clients: make(map[*websocket.Conn]models.ClientInfo),
	}
}

// project
func InitProjectWM() {
	WM3 = &models.WsProjectManager{
		ProjectInfo: make(map[*websocket.Conn]models.ProjectClientInfo),
	}
}
