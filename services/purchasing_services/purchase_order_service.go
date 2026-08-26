package purchasing_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/services/sales_services"
	"gorm.io/gorm"
)

type PurchaseOrderBody struct {
	models.PurchaseOrder
	PurchaseOrderDetails       []models.PurchaseOrderDetails `json:"purchase_order_details"`
	SalesOrderDetails          []models.OrderDetails         `json:"sales_order_details"`
	PurchaseRequisitionDetails []models.PROrders             `json:"purchase_requisition_details"`
}

func GetPurchaseOrder(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		PurchaseOrder        []models.PurchaseOrder                  `json:"purchaseorder"`
		PurchaseOrderDetails []models.PurchaseOrderDetailsWithStatus `json:"purchaseorderdetails"`
	}

	var response Response

	if err := services.DbGet(&response.PurchaseOrder, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting purchase order")
	}

	var plainDetails []models.PurchaseOrderDetails
	if err := GetPurchaseOrderDetails(&plainDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	receivedQtyByPodId, err := receivedQtyByPurchaseOrderDetailsId()
	if err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed computing received quantities")
	}

	response.PurchaseOrderDetails = make([]models.PurchaseOrderDetailsWithStatus, len(plainDetails))
	for i, d := range plainDetails {
		deliveryStatus := "WAITING FOR DELIVERY"
		if receivedQtyByPodId[d.ID] >= d.OrderQty {
			deliveryStatus = "IN STOCK"
		}
		response.PurchaseOrderDetails[i] = models.PurchaseOrderDetailsWithStatus{
			PurchaseOrderDetails: d,
			DeliveryStatus:       deliveryStatus,
		}
	}

	return response, 0, nil
}

// receivedQtyByPurchaseOrderDetailsId sums tbl_inv_receiving_report_details.received_qty
// per PO detail line - the same source sp_RecomputeSoItemStatus's own PO/RR block
// (§7.1 rows 3-4) uses, kept consistent rather than reusing the older
// tbl_inv_warehouse_receiving_history table a different, unrelated read path uses.
func receivedQtyByPurchaseOrderDetailsId() (map[uint]int, error) {
	type row struct {
		PurchaseOrderDetailsId uint
		ReceivedQty            int
	}
	var rows []row

	if err := initializers.DB.Raw(`
		SELECT purchase_order_details_id AS purchase_order_details_id, SUM(ISNULL(received_qty, 0)) AS received_qty
		FROM tbl_inv_receiving_report_details
		GROUP BY purchase_order_details_id
	`).Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[uint]int, len(rows))
	for _, r := range rows {
		result[r.PurchaseOrderDetailsId] = r.ReceivedQty
	}
	return result, nil
}

func CreatePurchaseOrder(c *fiber.Ctx, tx *gorm.DB) (PurchaseOrderBody, int, error) {
	var body PurchaseOrderBody
	if err := c.BodyParser(&body); err != nil {
		fmt.Println("ERROR", err)
		fmt.Println("PO INSERT BODY", err)
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.PurchaseOrder); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating purchaseorder")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	purchaseorderat := models.PurchaseOrderAt{
		RefId:                body.ID,
		PurchaseOrderContent: body.PurchaseOrderContent,
		At:                   at,
	}
	if err := services.DbInsert(tx, &purchaseorderat); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating purchaseorderat")
	}

	for _, v := range body.PurchaseOrderDetails {
		if err := CreatePurchaseOrderDetails(tx, body.ID, v, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	switch body.OrderType {
	case "SO":
		for _, detail := range body.SalesOrderDetails {
			conditions := map[string]interface{}{
				"order_details_id": detail.OrderDetailsID,
			}
			fmt.Println("PAPASOK SA UPDATE SOWD")
			if err := sales_services.UpdateSalesOrderDetails(tx, detail, at, conditions, body.OrderType, body.Status, "create"); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		}
	case "PR":
		for _, detail := range body.PurchaseRequisitionDetails {
			conditions := map[string]interface{}{
				"pr_order_id": detail.PR_Order_ID,
			}
			if err := UpdateRequisitionDetails(tx, detail, at, conditions, body.OrderType, body.Status, "create"); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		}
	default:
		return body, fiber.StatusBadRequest, errors.New("invalid order type")
	}

	InvalidatePOCaches()

	return body, 0, nil
}

func UpdatePurchaseOrder(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (PurchaseOrderBody, int, error) {
	var body PurchaseOrderBody

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body.PurchaseOrder, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating purchase order")
	}

	at, ok := c.Locals("at").(models.At)

	if !ok {
		at = models.At{}
	}

	atdata := models.PurchaseOrderAt{
		RefId:                body.ID,
		PurchaseOrderContent: body.PurchaseOrderContent,
		At:                   at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating purchaseorderat")
	}

	for _, v := range body.PurchaseOrderDetails {
		conditions := map[string]interface{}{
			"based_id": body.ID,
		}
		if err := UpdatePurchaseOrderDetails(tx, body.ID, v, at, conditions); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	switch body.OrderType {
	case "SO":
		fmt.Println("ORDER TYPE: SO UPDATEEEEEEEEEEEEEEEEEEEEEEE")
		for _, detail := range body.SalesOrderDetails {
			conditions := map[string]interface{}{
				"order_details_id": detail.OrderDetailsID,
			}
			fmt.Println("ORDER TYPE: SO UPDATE2222222222222222222")
			if err := sales_services.UpdateSalesOrderDetails(tx, detail, at, conditions, body.OrderType, body.Status, "update"); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		}
	case "PR":
		fmt.Println("ORDER TYPE: PR")
		for _, detail := range body.PurchaseRequisitionDetails {
			conditions := map[string]interface{}{
				"pr_order_id": detail.PR_Order_ID,
			}
			if err := UpdateRequisitionDetails(tx, detail, at, conditions, body.OrderType, body.Status, "update"); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		}
	default:
		return body, fiber.StatusBadRequest, errors.New("invalid order type")
	}

	InvalidatePOCaches()

	return body, fiber.StatusInternalServerError, nil
}

func InvalidatePOCaches() {
	cacheKeys := []interface{}{
		models.SOPurchasingListView{},
		models.PRPurchasingListView{},
		models.PurchasingListSupplierView{},
		models.PurchasingRedboxPurchaseListView{},
		models.PurchasingGuidingPriceView{},
	}
	for _, key := range cacheKeys {
		services.InvalidateCache(services.GetKey(key, nil))
	}
}
