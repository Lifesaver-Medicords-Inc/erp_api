package sales_handlers

import (
	"encoding/json"
	"fmt"
	"strconv"

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

// GetProjectHeadersPaged serves the project list: one page of latest project
// headers, newest first, no children. ?after= / ?before= / ?at= are ids.
func GetProjectHeadersPaged(c *fiber.Ctx) error {
	after, _ := strconv.Atoi(c.Query("after", "0"))
	before, _ := strconv.Atoi(c.Query("before", "0"))
	at, _ := strconv.Atoi(c.Query("at", "0"))

	data, pagination, status, err := sales_services.GetProjectHeadersPaged(after, before, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data, pagination)
}

// GetProjectDetail serves one project's full tree for lazy loading on open.
func GetProjectDetail(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "invalid project id")
	}

	data, status, svcErr := sales_services.GetProjectDetail(id)
	if svcErr != nil {
		return utils.RespondError(c, status, svcErr.Error())
	}
	return utils.RespondSuccess(c, data)
}

// GetProjectVersions serves one project document's full version history.
func GetProjectVersions(c *fiber.Ctx) error {
	data, status, err := sales_services.GetProjectVersions(c.Query("document_no"))
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data)
}
// SearchProjects backs the project Search: flat matching headers, 20 to a
// page, no children.
func SearchProjects(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))

	data, pagination, status, err := sales_services.SearchProjects(c.Query("term"), page)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}
	return utils.RespondSuccess(c, data, pagination)
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

	// Let anyone else with this same project quotation open know a save just happened, so
	// their client can refresh instead of waiting for its next 5-minute poll. Only fired once
	// the transaction has actually committed, so a notified client re-reading now always sees
	// this save, never a half-applied or rolled-back one. Scoped to this project's own WS
	// channel (see BroadcastToProject/WM3), so only viewers of this exact record get notified.
	// Best-effort: the save itself already succeeded above, so a broadcast failure (e.g. no
	// one currently connected) is only logged, never turned into a failed response.
	if err := services.BroadcastToProject("Sales", fmt.Sprint(data.ID), map[string]interface{}{
		"event": "quotation_saved",
		"id":    data.ID,
	}); err != nil {
		fmt.Println("Error broadcasting project update:", err)
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

// func CreateProjectItem(c *fiber.Ctx) error {
// 	tx := initializers.DB.Begin()
// 	if tx.Error != nil {
// 		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
// 	}

// 	data, status, err := sales_services.CreateNewProjectItemss(c, tx)

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

	// Same reasoning as UpdateProjectContent below - Advanced Conditions saves through its
	// own endpoint too, separately from the main "Save" button, and
	// data.SalesProjectContentAdvancedCondition.BasedId is the item set/tab id, not the
	// quotation id, so resolve the real quotation id through the item set before notifying.
	var itemSet models.SalesProjectItemSet
	if err := initializers.DB.Where("item_set_id = ?", data.SalesProjectContentAdvancedCondition.BasedId).First(&itemSet).Error; err == nil {
		if err := services.BroadcastToProject("Sales", fmt.Sprint(itemSet.BasedId), map[string]interface{}{
			"event": "quotation_saved",
			"id":    itemSet.BasedId,
		}); err != nil {
			fmt.Println("Error broadcasting project condition update:", err)
		}
	} else {
		fmt.Println("Error resolving quotation id for condition broadcast:", err)
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

	// Content fields (Flow/Head/Voltage/RPM/HP/Phase/etc.) save through this endpoint
	// separately from the main "Save" button (UpdateSalesProject above) - without this, an
	// edit made through this path never notified other viewers at all. data.BasedId here is
	// the item set/tab id, not the quotation id (see the "SHOULD BE THE TAB # / SET #" note
	// on SalesProjectContentContent.BasedId) - resolve the actual quotation id through the
	// item set so the broadcast reaches viewers of the right project. Queried against
	// initializers.DB (not tx - the transaction is already committed and done at this point),
	// after the commit above, for the same reason as UpdateSalesProject's broadcast.
	var itemSet models.SalesProjectItemSet
	if err := initializers.DB.Where("item_set_id = ?", data.BasedId).First(&itemSet).Error; err == nil {
		if err := services.BroadcastToProject("Sales", fmt.Sprint(itemSet.BasedId), map[string]interface{}{
			"event": "quotation_saved",
			"id":    itemSet.BasedId,
		}); err != nil {
			fmt.Println("Error broadcasting project content update:", err)
		}
	} else {
		fmt.Println("Error resolving quotation id for content broadcast:", err)
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

	// fmt.Println("Status:", status)

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

		var broadcastMsg = message

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
