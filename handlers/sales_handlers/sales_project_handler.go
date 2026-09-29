package sales_handlers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/handlers/engineering_handlers"
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

// GetItemPumps serves the pump electrical specs (FLA, VOLTAGE, ...) behind a project
// quotation's FINAL selection. ?ids=58,77 narrows it to those items: FINAL needs the
// specs of the pumps just picked, not of every pump in the catalogue, and it asks after
// the pick rather than before - the whole view is 4,884 rows on the rehearsal database
// and it used to be fetched before the picker could even open.
//
// No ids means the whole view, exactly as before.
func GetItemPumps(c *fiber.Ctx) error {
	var conditions map[string]interface{}
	if ids := utils.QueryIDs(c, "ids"); len(ids) > 0 {
		conditions = map[string]interface{}{"item_id": ids}
	}

	data, status, err := sales_services.GetItemPumps(conditions)
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

	// A project quote can be created with REQUEST FOR ENGR. already ticked - the flag
	// rides along in this same insert (see CreateSalesQuotation's own note on
	// is_requested_for_engr), so the engineer's list is out of date the moment this
	// commits. Same reasoning as the update path below.
	refreshEngineeringQuotationList()

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

	// REQUEST FOR ENGR. reaches the database through this save and nothing else: the
	// checkbox's own handler in Quotation.cs is commented out, so the standalone
	// RequestQuotationForEngr / CancelQuotationForEngr endpoints - which DO invalidate
	// this cache - are never called. Without this, a quotation sent to engineering was
	// written correctly but stayed invisible on the engineer's Sales Quotation List until
	// the cached read expired an hour later or the API restarted (user-reported 2026-09-24,
	// Q#0026). The same applies to anything else the view shows: project name, client,
	// created_by, and whether the quote has wiring rows.
	//
	// Best-effort and post-commit, for the same two reasons as the broadcast below.
	refreshEngineeringQuotationList()

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

// refreshEngineeringQuotationList brings every engineer's Sales Quotation List up to date
// after a project quotation is written. Two halves, and both are needed:
//
//   - the cache: the list is a cached read (EngineeringRedboxQuotationListView), so
//     without the invalidation anyone who opens or reloads the list keeps getting the
//     pre-save copy for up to an hour.
//   - the push: an engineer who already HAS the list open never asks again - the page
//     reads it once, on open. The engineering app keeps a socket on
//     /ws/engineering/redboxlist/quotation for its red box, and the API only ever sent on
//     it when a client connected. Sending the fresh list on it now is what makes a quote
//     requested in Sales appear in Engineering without anyone restarting or reopening
//     anything (user-reported 2026-09-24).
//
// Must run AFTER tx.Commit(): the push re-reads the list, and re-reading inside the
// transaction would publish (and re-cache) the pre-commit state. The invalidation is
// synchronous so the very next GET is fresh; the push runs in the background so a slow or
// dead engineering socket can never hold up the save's own response. Best-effort
// throughout - the save has already succeeded by the time either runs.
func refreshEngineeringQuotationList() {
	if err := sales_services.InvalidateEngineeringQuotationListCache(); err != nil {
		fmt.Println("Error invalidating engineering quotation list cache:", err)
	}

	go func() {
		if err := engineering_handlers.BroadcastRedboxQuotationList(); err != nil {
			fmt.Println("Error broadcasting engineering quotation list:", err)
		}
	}()
}
