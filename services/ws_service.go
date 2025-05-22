package services

import (
	"errors"
	"fmt"

	"github.com/gofiber/contrib/websocket"
	"github.com/pierceperado/smpc/initializers"
)

func HandleWs(c *websocket.Conn, h func(*websocket.Conn)) {
	initializers.WM.AddClient(c)
	fmt.Println("Client Connected:", c.IP())

	defer func() {
		initializers.WM.RemoveClient(c)
		fmt.Println("Client Disconnected:", c.IP())
	}()

	h(c)

	for {
		var message interface{}
		if err := c.ReadJSON(&message); err != nil {
			fmt.Println("error reading message")
			break
		}

		if err := c.WriteJSON(message); err != nil {
			fmt.Println("error writing message")
			break
		}
	}
}

func BroadcastMessage(data interface{}) error {
	initializers.WM.RLock()
	defer initializers.WM.RUnlock()

	for client := range initializers.WM.Clients {
		if err := client.WriteJSON(data); err != nil {
			return errors.New("error sending message")
		}
	}

	return nil
}

// for deparments redbox
func HandleWs2(client *websocket.Conn, handler func(*websocket.Conn, string)) {
	department := client.Query("department")
	initializers.WM2.AddClient2(client, department)
	fmt.Println("Client Connected:", client.IP(), department)

	defer func() {
		initializers.WM2.RemoveClient2(client)
		fmt.Println("Client Disconnected:", client.IP(), department)
	}()

	handler(client, department)
}

// entry point
func HandleProjectWs(client *websocket.Conn, handler func(*websocket.Conn, string, string)) {
	branch := client.Query("branch")
	projectid := client.Query("projectid")
	initializers.WM3.AddProjectClient(client, branch, projectid)
	fmt.Println("Client Connected:", client.IP(), "BRANCH: ", branch, "PROJECT ID: ", projectid)

	defer func() {
		initializers.WM3.RemoveProjectClient(client)
		fmt.Println("Client Disconnected:", client.IP(), "BRANCH: ", branch, "PROJECT ID: ", projectid)
	}()

	handler(client, branch, projectid)
}

func BroadcastToDepartment(dept string, data interface{}) error {
	initializers.WM2.RLock()
	defer initializers.WM2.RUnlock()

	for _, info := range initializers.WM2.Clients {
		if info.Conn == nil {
			fmt.Println("WARNING: Conn is nil for department", info.Department)
			continue
		}

		if info.Department == dept {
			if err := info.Conn.WriteJSON(data); err != nil {
				fmt.Println("HELLOWORLD:", err)
				return err
			}
		}
	}

	return nil
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func BroadcastToDepartments(depts []string, data interface{}) error {
	initializers.WM2.RLock()
	defer initializers.WM2.RUnlock()

	for _, info := range initializers.WM2.Clients {
		if info.Conn == nil {
			fmt.Println("WARNING: Conn is nil for department", info.Department)
			continue
		}

		if contains(depts, info.Department) {
			if err := info.Conn.WriteJSON(data); err != nil {
				fmt.Println("HELLOWORLD:", err)
				return err
			}
		}
	}

	return nil
}

func BroadcastToProject(branch string, projectid string, data interface{}) error {
	initializers.WM3.RLock()
	defer initializers.WM3.RUnlock()

	for _, info := range initializers.WM3.ProjectInfo {
		if info.Conn == nil {
			fmt.Println("WARNING: Conn is nil for department", info.Branch)
			continue
		}

		if info.Branch == branch && info.ProjectId == projectid {
			if err := info.Conn.WriteJSON(data); err != nil {
				fmt.Println("HELLOWORLD:", err)
				return err
			}
		}
	}

	return nil
}
