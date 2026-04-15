package purchasing_services

import (
	"errors"
	"fmt"

	// "fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetPROrders(prorders *[]models.PROrders, conditions map[string]interface{}) error {
	if err := services.DbGet(prorders, conditions); err != nil {
		return errors.New("failed getting purchase requisition orders")
	}

	return nil
}

// func GetOrderDetails(conditions map[string]interface{}) ([]models.OrderDetails, int, error) {
// 	var orders []models.OrderDetails

// 	if err := services.DbGet(&orders, conditions); err != nil {
// 		return orders, fiber.StatusInternalServerError, errors.New("failed getting order details")
// 	}

//		return orders, 0, nil
//	}

func GetPROrder(prorders *models.PROrders, conditions map[string]interface{}) error {
	if err := services.DbGet(prorders, conditions); err != nil {
		return errors.New("failed getting purchase requisition orders")
	}
	return nil
}

func CreatePROrder(tx *gorm.DB, parentId uint, PROrder models.PROrders, at models.At) error {
	// parentId = 10015
	PROrder.Based_ID = parentId

	if err := services.DbInsert(tx, &PROrder); err != nil {
		return errors.New("failed creating purchase requisition orders")
	}

	prordersat := models.PROrdersAt{
		RefId:           PROrder.PR_Order_ID,
		PROrdersContent: PROrder.PROrdersContent,
		At:              at,
	}

	if err := services.DbInsert(tx, &prordersat); err != nil {
		return errors.New("failed creating purchase requisition orders")
	}

	return nil
}

func UpdatePROrder(tx *gorm.DB, prorders models.PROrders, at models.At, conditions map[string]interface{}) error {
	// Ensure there is a condition to update the specific order details
	conditions = map[string]interface{}{
		"pr_order_id": prorders.PR_Order_ID,
	}
	// Perform the update for order details with a condition
	if err := services.DbUpdate(tx, &prorders, conditions); err != nil {
		return errors.New("failed updating purchase requisition orders")
	}

	// Insert the "at" data for the updated order details
	prorderat := models.PROrdersAt{
		RefId:           prorders.PR_Order_ID,
		PROrdersContent: prorders.PROrdersContent,
		At:              at,
	}

	if err := services.DbInsert(tx, &prorderat); err != nil {
		return errors.New("failed creating purchase requisition orders at")
	}

	InvalidatePRCaches()
	return nil
}

func UpdateRequisitionDetails(tx *gorm.DB, orderdetails models.PROrders, at models.At, conditions map[string]interface{}, orderType string, status string, mode string) error {
	if len(conditions) == 0 {
		conditions = map[string]interface{}{
			"pr_order_id": orderdetails.PR_Order_ID,
		}
	}
	fmt.Println("PR UPDATE DETAILS")
	var existing models.PROrders

	if err := tx.Model(&models.PROrders{}).Where(conditions).First(&existing).Error; err != nil {
		return errors.New("failed getting existing requisition order")
	}

	if status == "CANCELLED" && mode == "update" {
		fmt.Println("ORDER DATA: ", orderdetails)

		// Defensive checks in case pointers are nil
		if existing.AllocatedQty == nil {
			existing.AllocatedQty = new(int)
		}
		if orderdetails.AllocatedQty == nil {
			orderdetails.AllocatedQty = new(int)
		}

		result := *existing.AllocatedQty - *orderdetails.AllocatedQty
		if result < 0 {
			result = 0
		}

		fmt.Printf("EXISTING ALLOC: %d, CANCELLED: %d, RESULT: %d\n",
			*existing.AllocatedQty,
			*orderdetails.AllocatedQty,
			result)

		fmt.Println("PO CANCELLED, DEDUCT ALLOC QTY")

		*orderdetails.AllocatedQty = result
	} else if mode == "create" {
		fmt.Println("CREATE MODE")

		if orderdetails.AllocatedQty == nil {
			orderdetails.AllocatedQty = new(int)
		}
		if existing.AllocatedQty != nil {
			*orderdetails.AllocatedQty += *existing.AllocatedQty
		}
	}

	if err := services.DbUpdate(tx, &orderdetails, conditions); err != nil {
		return errors.New("failed updating requisition order")
	}

	if err := tx.Exec("EXEC sp_SetOrderStatus ?, ?", orderdetails.PR_Order_ID, orderType).Error; err != nil {
		return errors.New("failed executing stored procedure")
	}

	orderdetailsat := models.PROrdersAt{
		RefId:           orderdetails.PR_Order_ID,
		PROrdersContent: orderdetails.PROrdersContent,
		At:              at,
	}

	if err := services.DbInsert(tx, &orderdetailsat); err != nil {
		return errors.New("failed creating requisition orders at")
	}

	return nil
}

func DeletePROrder(tx *gorm.DB, prorder models.PROrders, at models.At, conditions map[string]interface{}) error {
	if err := services.DbDelete(tx, &prorder, conditions); err != nil {
		return errors.New("failed deleting purchase requisition orders")
	}

	prorderat := models.PROrdersAt{
		RefId:           prorder.PR_Order_ID,
		PROrdersContent: prorder.PROrdersContent,
		At:              at,
	}
	if err := services.DbInsert(tx, &prorderat); err != nil {
		return errors.New("failed deleting purchase requisition orders at")
	}

	return nil
}

func InvalidatePRCaches() {
	cacheKeys := []interface{}{
		models.PRPurchasingListView{},
		models.PurchasingListSupplierView{},
		models.PurchasingRedboxPurchaseListView{},
	}
	for _, key := range cacheKeys {
		services.InvalidateCache(services.GetKey(key, nil))
	}
}
