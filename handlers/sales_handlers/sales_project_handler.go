package sales_handlers

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/sales_services"

	"github.com/pierceperado/smpc/utils"
)

func GetSalesProject(c *fiber.Ctx) error {
	data, status, err := sales_services.GetSalesProjects(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func GetSalesCanvasView(c *fiber.Ctx) error {
	data, status, err := sales_services.GetSalesCanvasView(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func GetItemPumps(c *fiber.Ctx) error {
	data, status, err := sales_services.GetItemPumps(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func GetBpiSuppliers(c *fiber.Ctx) error {
	data, status, err := sales_services.GetBpiSuppliers(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}

func CreateSalesProject(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := sales_services.CreateSalesProject(c, tx)
	fmt.Println("DATAAAA", data)

	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func UpdateSalesProject(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := sales_services.UpdateSalesProject(c, tx)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func CreateItemSetTab(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}
	data, status, err := sales_services.CreateNewItems(c, tx)

	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func CreateProjectItem(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := sales_services.CreateNewProjectItemss(c, tx)

	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func CreateProjectWirings(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	data, status, err := sales_services.CreateNewProjectWiring(c, tx)

	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

// func UpdateProjectMultiplier(c *fiber.Ctx) error {
// 	tx := initializers.DB.Begin()
// 	if tx.Error != nil {
// 		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
// 	}

// 	data, status, err := sales_services.UpdateSalesProjectMultiplier()
// 	if err != nil {
// 		tx.Rollback()
// 		return utils.RespondError(c, status, err.Error())
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		tx.Rollback()
// 		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
// 	}

// 	return utils.RespondSuccess(c, data)
// }

func UpdateProjectItem(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := sales_services.UpdateProjectItemss(c, tx, nil)

	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func UpdateProjectWiring(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := sales_services.UpdateProjectWirings(c, tx, nil)

	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

type departments struct {
	Departments []string `json:"departments"`
}

func UpdateProjectCondition(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}
	data, status, err := sales_services.UpdateProjectAdvancedCondition(c, tx, nil)

	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func UpdateProjectContent(c *fiber.Ctx) error {

	var body models.SalesProjectContent
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := sales_services.UpdateProjectContents(c, tx, nil)

	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func CreateSalesCanvasSheet(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := sales_services.CreateSalesCanvasSheet(c, tx)

	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func BroadcastProjects(department departments) error {
	data, status, err := sales_services.GetSalesProjects(nil)
	if err != nil {
		return err
	}

	fmt.Println("Status:", status)

	if err := services.BroadcastToDepartments(department.Departments, data); err != nil {
		return err
	}
	return nil
}

func WsProjects(c *websocket.Conn, userid string, branch string, projectid string) {
	// fmt.Println("WsProjects: ", branch, projectid)

	// projectconditions := map[string]interface{}{
	// 	"id": projectid,
	// }

	// multiplierconditions := map[string]interface{}{
	// 	"based_id": projectid,
	// }

	// data, status, err := sales_services.GetSalesProjectsWS(projectconditions, multiplierconditions)
	// if err != nil {
	// 	fmt.Println("Error getting initial data:", err)
	// 	return
	// }

	// if err := services.BroadcastToProject(branch, projectid, data); err != nil {
	// 	fmt.Println("Error broadcasting initial data:", err)
	// 	return
	// }

	//fmt.Println("Status:", status)

	for {
		_, msgBytes, err := c.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		fmt.Println("Received raw message:", string(msgBytes))

		var message interface{}

		if err := json.Unmarshal(msgBytes, &message); err != nil {
			fmt.Println("Invalid JSON received:", err)
			continue
		}

		var broadcastMsg interface{} = message

		if msgMap, ok := message.(map[string]interface{}); ok {
			// Remove "projectid" and "branch" keys if present
			delete(msgMap, "project_id")
			delete(msgMap, "branch")
			broadcastMsg = msgMap
		}

		if err := services.BroadcastToProject(branch, projectid, broadcastMsg); err != nil {
			fmt.Println("Error broadcasting initial data:", err)
			return
		}

	}
}
