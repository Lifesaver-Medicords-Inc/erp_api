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
	fmt.Println("starting: ", PRbody.PROrder)
	if err := c.BodyParser(&PRbody.PROrder); err != nil {
		fmt.Println(err)
		return PRbody, fiber.StatusBadRequest, errors.New("cannot bind request")
	}
	fmt.Println("after parser: ", PRbody.PROrder)
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	if err := CreatePROrder(tx, PRbody.PROrder.Based_ID, PRbody.PROrder, at); err != nil {
		fmt.Println(PRbody)
		return PRbody, fiber.StatusInternalServerError, err
	}
	fmt.Println(PRbody)
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

	conditions = map[string]interface{}{
		"doc_no": bodyPR.DocNo,
	}

	if err := services.DbUpdate(tx, &bodyPR.PurchaseRequisition, conditions); err != nil {
		fmt.Println(err)
		return bodyPR, fiber.StatusInternalServerError, errors.New("failed updating purchase requisition")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	for _, PRorder := range bodyPR.PROrder {
		var existingOrderDetail models.PROrders
		proOrderConditions := map[string]interface{}{
			"pr_order_id": PRorder.PR_Order_ID,
		}

		if err := services.DbGet(&existingOrderDetail, proOrderConditions); err != nil {
			PRorder.Based_ID = bodyPR.PurchaseRequisition.PR_ID
			if err := CreatePROrder(tx, bodyPR.PurchaseRequisition.PR_ID, PRorder, at); err != nil {
				return bodyPR, fiber.StatusInternalServerError, err
			}
		} else {
			PRorder.Based_ID = bodyPR.PurchaseRequisition.PR_ID
			if err := UpdatePROrder(tx, PRorder, at, proOrderConditions); err != nil {
				return bodyPR, fiber.StatusInternalServerError, err
			}
		}
	}
	fmt.Println(bodyPR)

	redboxpurchasinglistview := services.GetKey(models.PurchasingRedboxPurchaseListView{}, nil)
	services.InvalidateCache(redboxpurchasinglistview)

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

func DeletePROrderByID(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (BodyPR, int, error) {
	var bodyPR BodyPR
	if err := c.BodyParser(&bodyPR.PROrder); err != nil {
		return bodyPR, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	conditions = map[string]interface{}{
		"pr_order_id": bodyPR.PROrder.PR_Order_ID,
	}

	if err := DeletePROrder(tx, bodyPR.PROrder, at, conditions); err != nil {
		return bodyPR, fiber.StatusInternalServerError, err
	}
	fmt.Println(bodyPR)
	return bodyPR, 0, nil
}
