package sales_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type BodyOrder struct {
	models.Order
	OrderDetails models.OrderDetails `json:"sales_order_details"`
}

type BodyOrderDetails struct {
	models.Order
	//Child 1
	OrderDetails []models.OrderDetails `json:"sales_order_details"`
}

func GetOrders(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		Orders       []models.Order        `json:"order"`
		OrderDetails []models.OrderDetails `json:"sales_order_details"`
	}

	var response Response
	fmt.Println("Orders: ", response)
	if err := services.DbGet(&response.Orders, conditions); err != nil {
		fmt.Println(err)
		return response, fiber.StatusInternalServerError, errors.New("failed getting orders")
	}

	if err := GetOrderDetails(&response.OrderDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	return response, 0, nil
}

func GetOrder(Order_ID int) (BodyOrder, int, error) {
	conditions := map[string]interface{}{
		"order_id": Order_ID,
	}

	var orderrecord BodyOrder

	if err := services.DbGet(&orderrecord.Order, conditions); err != nil {
		return orderrecord, fiber.StatusInternalServerError, errors.New("failed getting order")
	}

	conditions = map[string]interface{}{
		// based on parent ID
		"based_id": orderrecord.Order.Order_ID,
	}

	//Child 1
	if err := GetOrderDetail(&orderrecord.OrderDetails, conditions); err != nil {
		return orderrecord, fiber.StatusInternalServerError, err
	}

	// //Child 2 for project quote
	// if err := GetSalesQuotationQuick(&record.QuickQuote, conditions); err != nil {
	// 	return record, fiber.StatusInternalServerError, err
	// }

	return orderrecord, 0, nil
}

// func GetOrder(Order_ID int) (Body, int, error) {
// 	conditions := map[string]interface{}{
// 		"order_id": Order_ID,
// 	}

// 	var record Body

// 	if err := services.DbGet(&record.Order, conditions); err != nil {
// 		return record, fiber.StatusInternalServerError, errors.New("failed getting order")
// 	}

// 	conditions = map[string]interface{}{
// 		// based on parent ID
// 		"based_id": record.Order.Order_ID,
// 	}

// 	//Child 1
// 	if err := GetOrderDetails(&record.OrderDetails, conditions); err != nil {
// 		return record, fiber.StatusInternalServerError, err
// 	}

// 	// //Child 2 for project quote
// 	// if err := GetSalesQuotationQuick(&record.QuickQuote, conditions); err != nil {
// 	// 	return record, fiber.StatusInternalServerError, err
// 	// }

// 	return record, 0, nil
// }

// CREATE CHILD SERVICE
func CreateOrderChild(c *fiber.Ctx, tx *gorm.DB) (BodyOrder, int, error) {
	var orderbody BodyOrder
	if err := c.BodyParser(&orderbody); err != nil {
		return orderbody, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	if err := CreateOrderDetail(tx, orderbody.Order_ID, orderbody.OrderDetails, at); err != nil {
		return orderbody, fiber.StatusInternalServerError, err
	}

	return orderbody, 0, nil
}

func CreateOrder(c *fiber.Ctx, tx *gorm.DB) (BodyOrderDetails, int, error) {
	var bodyorder BodyOrderDetails
	if err := c.BodyParser(&bodyorder); err != nil {
		fmt.Println(err)
		return bodyorder, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &bodyorder.Order); err != nil {
		return bodyorder, fiber.StatusInternalServerError, errors.New("failed creating order")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.OrderAt{RefId: bodyorder.Order_ID, OrderContent: bodyorder.OrderContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return bodyorder, fiber.StatusInternalServerError, errors.New("failed creating order at")
	}

	for _, v := range bodyorder.OrderDetails {
		if err := CreateOrderDetail(tx, bodyorder.Order_ID, v, at); err != nil {
			return bodyorder, fiber.StatusInternalServerError, err
		}
	}
	// if err := CreateSalesQuotationQuick(tx, body.ID, body.SalesQuotationQuick[], at); err != nil {
	// 	return body, fiber.StatusInternalServerError, err
	// }
	return bodyorder, 0, nil
}

// UpdateOrder function or similar code
func UpdateOrder(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (BodyOrderDetails, int, error) {
	var bodyorder BodyOrderDetails
	if err := c.BodyParser(&bodyorder); err != nil {
		fmt.Println(err)
		return bodyorder, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	// Update the parent order (already done in your existing code)
	conditions = map[string]interface{}{
		"doc": bodyorder.Doc,
	}

	if err := services.DbUpdate(tx, &bodyorder.Order, conditions); err != nil {
		fmt.Println(err)
		return bodyorder, fiber.StatusInternalServerError, errors.New("failed updating order")
	}

	// Now update each order detail (child) individually
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	// Iterate over the slice of order details and update each one
	for _, orderDetail := range bodyorder.OrderDetails {
		// Add a condition based on the order details ID (or other relevant fields)
		orderDetailConditions := map[string]interface{}{
			"based_id": bodyorder.Order_ID, // Assuming 'based_id' is the condition to match
		}

		// Call UpdateOrderDetail for each child (order detail)
		if err := UpdateOrderDetail(tx, orderDetail, at, orderDetailConditions); err != nil {
			return bodyorder, fiber.StatusInternalServerError, err
		}
		fmt.Println(bodyorder)
	}
	fmt.Println(bodyorder)
	// If everything goes well, return success

	purchasinglistview := services.GetKey(models.PurchasingListView{}, nil)
	services.InvalidateCache(purchasinglistview)

	purchasinsupplierlist := services.GetKey(models.PurchasingListSupplierView{}, nil)
	services.InvalidateCache(purchasinsupplierlist)

	return bodyorder, 0, nil
}

func DeleteOrder(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (BodyOrder, int, error) {
	var bodyorder BodyOrder
	if err := c.BodyParser(&bodyorder); err != nil {
		return bodyorder, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &bodyorder.Order, conditions); err != nil {
		return bodyorder, fiber.StatusInternalServerError, errors.New("failed deleting order")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.OrderAt{RefId: bodyorder.Order_ID, OrderContent: bodyorder.OrderContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return bodyorder, fiber.StatusInternalServerError, errors.New("failed creating order at")
	}

	conditions = map[string]interface{}{
		"based_id": bodyorder.Order_ID,
	}

	if err := DeleteOrderDetail(tx, bodyorder.OrderDetails, at, conditions); err != nil {
		return bodyorder, fiber.StatusInternalServerError, err
	}

	return bodyorder, 0, nil
}

// func DeleteOrder(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Order, int, error) {
// 	var body models.Order
// 	if err := c.BodyParser(&body); err != nil {
// 		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
// 	}

// 	if err := services.DbDelete(tx, &body, conditions); err != nil {
// 		return body, fiber.StatusInternalServerError, errors.New("failed deleting order")
// 	}

// 	at, ok := c.Locals("at").(models.At)
// 	if !ok {
// 		at = models.At{}
// 	}

// 	atdata := models.OrderAt{RefId: body.Order_ID, OrderContent: body.OrderContent, At: at}
// 	if err := services.DbInsert(tx, &atdata); err != nil {
// 		return body, fiber.StatusInternalServerError, errors.New("failed creating orderat")
// 	}

// 	return body, 0, nil
// }
