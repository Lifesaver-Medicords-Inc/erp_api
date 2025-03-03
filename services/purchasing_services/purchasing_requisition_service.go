package purchasing_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type BodyPR struct {
	models.PurchaseRequisition
	PROrder models.PROrders `json:"purchasing_purchase_requisition"`
}

type BodyPROrder struct {
	models.PurchaseRequisition
	//Child 1
	PROrder []models.PROrders `json:"purchasing_purchase_requisition_orders"`
}

func GetPRs(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		PRs      []models.PurchaseRequisition `json:"purchase_requisition"`
		PROrders []models.PROrders            `json:"purchasing_purchase_requisition_orders"`
	}

	var response Response
	fmt.Println("PRs: ", response)
	if err := services.DbGet(&response.PRs, conditions); err != nil {
		fmt.Println(err)
		return response, fiber.StatusInternalServerError, errors.New("failed getting purchase requisitions")
	}

	if err := GetPROrders(&response.PROrders, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	return response, 0, nil
}

func GetPR(PR_ID int) (BodyPR, int, error) {
	conditions := map[string]interface{}{
		"pr_id": PR_ID,
	}

	var PRrecord BodyPR

	if err := services.DbGet(&PRrecord.PurchaseRequisition, conditions); err != nil {
		return PRrecord, fiber.StatusInternalServerError, errors.New("failed getting purchase requisition")
	}

	conditions = map[string]interface{}{
		// based on parent ID
		"based_id": PRrecord.PurchaseRequisition.PR_ID,
	}

	//Child 1
	if err := GetPROrder(&PRrecord.PROrder, conditions); err != nil {
		return PRrecord, fiber.StatusInternalServerError, err
	}

	return PRrecord, 0, nil
}

// CREATE CHILD SERVICE
func CreatePRChild(c *fiber.Ctx, tx *gorm.DB) (BodyPR, int, error) {
	var PRbody BodyPR
	if err := c.BodyParser(&PRbody); err != nil {
		return PRbody, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	if err := CreatePROrder(tx, PRbody.PR_ID, PRbody.PROrder, at); err != nil {
		return PRbody, fiber.StatusInternalServerError, err
	}

	return PRbody, 0, nil
}

func CreatePR(c *fiber.Ctx, tx *gorm.DB) (BodyPROrder, int, error) {
	var bodyPR BodyPROrder
	if err := c.BodyParser(&bodyPR); err != nil {
		fmt.Println(err)
		return bodyPR, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &bodyPR.PurchaseRequisition); err != nil {
		return bodyPR, fiber.StatusInternalServerError, errors.New("failed creating purchase requisition")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PurchaseRequisitionAt{RefId: bodyPR.PR_ID, PurchaseRequisitionContent: bodyPR.PurchaseRequisitionContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return bodyPR, fiber.StatusInternalServerError, errors.New("failed creating purchase requisition at")
	}

	for _, v := range bodyPR.PROrder {
		if err := CreatePROrder(tx, bodyPR.PR_ID, v, at); err != nil {
			return bodyPR, fiber.StatusInternalServerError, err
		}
	}
	// if err := CreateSalesQuotationQuick(tx, body.ID, body.SalesQuotationQuick[], at); err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }
	return bodyPR, 0, nil
}

// UpdateOrder function or similar code
func UpdatePR(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (BodyPROrder, int, error) {
	var bodyPR BodyPROrder
	if err := c.BodyParser(&bodyPR); err != nil {
		fmt.Println(err)
		return bodyPR, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	// Update the parent order (already done in your existing code)
	conditions = map[string]interface{}{
		"doc_no": bodyPR.DocNo,
	}

	if err := services.DbUpdate(tx, &bodyPR.PurchaseRequisition, conditions); err != nil {
		fmt.Println(err)
		return bodyPR, fiber.StatusInternalServerError, errors.New("failed updating purchase requisition")
	}

	// Now update each order detail (child) individually
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	// Iterate over the slice of order details and update each one
	for _, PRorder := range bodyPR.PROrder {
		// Add a condition based on the order details ID (or other relevant fields)
		PROrderConditions := map[string]interface{}{
			"based_id": bodyPR.PR_ID, // Assuming 'based_id' is the condition to match
		}

		// Call UpdateOrderDetail for each child (order detail)
		if err := UpdatePROrder(tx, PRorder, at, PROrderConditions); err != nil {
			return bodyPR, fiber.StatusInternalServerError, err
		}
		fmt.Println(bodyPR)
	}
	fmt.Println(bodyPR)
	// If everything goes well, return success
	return bodyPR, 0, nil
}

func DeletePR(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (BodyPR, int, error) {
	var bodyPR BodyPR
	if err := c.BodyParser(&bodyPR); err != nil {
		return bodyPR, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &bodyPR.PurchaseRequisition, conditions); err != nil {
		return bodyPR, fiber.StatusInternalServerError, errors.New("failed deleting purchase requisition")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PurchaseRequisitionAt{RefId: bodyPR.PR_ID, PurchaseRequisitionContent: bodyPR.PurchaseRequisitionContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return bodyPR, fiber.StatusInternalServerError, errors.New("failed creating purchase requisition at")
	}

	conditions = map[string]interface{}{
		"based_id": bodyPR.PR_ID,
	}

	if err := DeletePROrder(tx, bodyPR.PROrder, at, conditions); err != nil {
		return bodyPR, fiber.StatusInternalServerError, err
	}

	return bodyPR, 0, nil
}
