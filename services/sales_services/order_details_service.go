package sales_services

import (
	"errors"
	// "fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetOrderDetails(orderdetails *[]models.OrderDetails, conditions map[string]interface{}) error {
	if err := services.DbGet(orderdetails, conditions); err != nil {
		return errors.New("failed getting order details")
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

func GetOrderDetail(orderdetails *models.OrderDetails, conditions map[string]interface{}) error {
	if err := services.DbGet(orderdetails, conditions); err != nil {
		return errors.New("failed getting order detail")
	}

	return nil
}

func CreateOrderDetail(tx *gorm.DB, parentId uint, OrderDetails models.OrderDetails, at models.At) error {

	// parentId = 10015
	OrderDetails.Based_ID = parentId

	if err := services.DbInsert(tx, &OrderDetails); err != nil {
		return errors.New("failed creating order details")
	}

	orderdetailsat := models.OrderDetailsAt{
		RefId:               OrderDetails.Order_Details_ID,
		OrderDetailsContent: OrderDetails.OrderDetailsContent,
		At:                  at,
	}

	if err := services.DbInsert(tx, &orderdetailsat); err != nil {
		return errors.New("failed creating quick quotations")
	}

	return nil
}

func UpdateOrderDetail(tx *gorm.DB, orderdetails models.OrderDetails, at models.At, conditions map[string]interface{}) error {
	// Ensure there is a condition to update the specific order details
	conditions = map[string]interface{}{
		"based_id": orderdetails.Based_ID,
	}
	// Perform the update for order details with a condition
	if err := services.DbUpdate(tx, &orderdetails, conditions); err != nil {
		return errors.New("failed updating order details")
	}

	// Insert the "at" data for the updated order details
	orderdetailsat := models.OrderDetailsAt{
		RefId:               orderdetails.Order_Details_ID,
		OrderDetailsContent: orderdetails.OrderDetailsContent,
		At:                  at,
	}

	if err := services.DbInsert(tx, &orderdetailsat); err != nil {
		return errors.New("failed creating orderdetailsat")
	}

	return nil
}

func DeleteOrderDetail(tx *gorm.DB, orderdetails models.OrderDetails, at models.At, conditions map[string]interface{}) error {
	if err := services.DbDelete(tx, &orderdetails, conditions); err != nil {
		return errors.New("failed deleting order details")
	}

	orderdetailsat := models.OrderDetailsAt{
		RefId:               orderdetails.Order_Details_ID,
		OrderDetailsContent: orderdetails.OrderDetailsContent,
		At:                  at,
	}
	if err := services.DbInsert(tx, &orderdetailsat); err != nil {
		return errors.New("failed deleting order detail at")
	}

	return nil
}
