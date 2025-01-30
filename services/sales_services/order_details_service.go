package sales_services

import (
	"errors"
	// "fmt"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetOrderDetails(quickquotes *[]models.OrderDetails, conditions map[string]interface{}) error {
	if err := services.DbGet(quickquotes, conditions); err != nil {
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

func CreateOrderDetail(tx *gorm.DB, basedId uint, OrderDetails models.OrderDetails, at models.At) error {
	content := models.OrderDetailsContent{
		Based_ID:           basedId,
		Qty:                OrderDetails.Qty,
		ItemCode:           OrderDetails.ItemCode,
		ItemDescription:    OrderDetails.ItemDescription,
		DeliveryPreference: OrderDetails.DeliveryPreference,
		ListPrice:          OrderDetails.ListPrice,
		TotalPrice:         OrderDetails.TotalPrice,
		Status:             OrderDetails.Status,
	}

	orderdetails := models.OrderDetails{OrderDetailsContent: content}
	if err := services.DbInsert(tx, &orderdetails); err != nil {
		return errors.New("failed creating order detail")
	}

	orderdetailsat := models.OrderDetailsAt{
		RefId:               orderdetails.Order_Details_ID,
		OrderDetailsContent: content,
		At:                  at,
	}

	if err := services.DbInsert(tx, &orderdetailsat); err != nil {
		return errors.New("failed creating order detail")
	}

	return nil
}

func UpdateOrderDetail(tx *gorm.DB, orderdetails models.OrderDetails, at models.At, conditions map[string]interface{}) error {

	if err := services.DbUpdate(tx, &orderdetails, nil); err != nil {
		return errors.New("failed updating order details")
	}

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
