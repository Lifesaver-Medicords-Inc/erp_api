package purchasing_services

import (
	"errors"
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

	return nil
}

func UpdateRequisitionDetails(tx *gorm.DB, orderdetails models.PROrders, at models.At, conditions map[string]interface{}) error {
	// If no conditions passed, fallback to primary key condition
	if len(conditions) == 0 {
		conditions = map[string]interface{}{
			"pr_order_id": orderdetails.PR_Order_ID,
		}
	}

	if err := services.DbUpdate(tx, &orderdetails, conditions); err != nil {
		return errors.New("failed updating requisition orders")
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
