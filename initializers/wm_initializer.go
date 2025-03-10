package initializers

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/pierceperado/smpc/models"
)

var WM *models.WsManager

func InitWm() {
	WM = &models.WsManager{
		Clients: make(map[*websocket.Conn]bool),
	}
}
