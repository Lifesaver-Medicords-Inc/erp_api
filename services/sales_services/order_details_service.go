package sales_services

import (
	"errors"
	"fmt"

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
		RefId:               OrderDetails.OrderDetailsID,
		OrderDetailsContent: OrderDetails.OrderDetailsContent,
		At:                  at,
	}

	if err := services.DbInsert(tx, &orderdetailsat); err != nil {
		return errors.New("failed creating quick quotations")
	}

	return nil
}

func UpdateOrderDetail(tx *gorm.DB, orderdetails models.OrderDetails, at models.At, conditions map[string]interface{}) error {
	conditions = map[string]interface{}{
		"order_details_id": orderdetails.OrderDetailsID,
		"based_id":         orderdetails.Based_ID,
	}

	if err := services.DbUpdate(tx, &orderdetails, conditions); err != nil {
		return errors.New("failed updating order details")
	}

	var full models.OrderDetails
	if err := tx.Where(conditions).First(&full).Error; err != nil {
		return errors.New("failed fetching updated order details")
	}

	orderdetailsat := models.OrderDetailsAt{
		RefId:               full.OrderDetailsID,
		OrderDetailsContent: full.OrderDetailsContent,
		At:                  at,
	}

	if err := services.DbInsert(tx, &orderdetailsat); err != nil {
		return errors.New("failed creating orderdetailsat")
	}

	return nil
}

func UpdateSalesOrderDetails(tx *gorm.DB, orderdetails models.OrderDetails, at models.At, conditions map[string]interface{}, orderType string, status string, mode string) error {
	fmt.Println("UPDATE SO SERVICE ORDER DETAILS", orderdetails)
	if len(conditions) == 0 {
		conditions = map[string]interface{}{
			"order_details_id": orderdetails.OrderDetailsID,
		}
	}

	var existing models.OrderDetails

	if err := tx.Model(&models.OrderDetails{}).Where(conditions).First(&existing).Error; err != nil {
		return errors.New("failed getting existing order details")
	}

	fmt.Printf("Fetched existing order details: %+v\n", existing)

	// This function is only ever called from the Purchasing module (creating/
	// cancelling a PO against a sales order line - see purchase_order_service.go),
	// whose request payload has no concept of item_set_header at all. DbUpdate
	// below does a full-column UpdateColumns, which (unlike GORM's Updates)
	// writes every field including zero values - so without this, every PO
	// create/update against a line item would silently blank out the itemset
	// header label captured when the sales order was saved. Purchasing should
	// never be able to touch this field, so always carry the existing value
	// forward regardless of what's on the incoming struct.
	orderdetails.ItemSetHeader = existing.ItemSetHeader

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

		// Ensure both pointers are initialized
		if existing.AllocatedQty == nil {
			existing.AllocatedQty = new(int)
		}
		if orderdetails.AllocatedQty == nil {
			orderdetails.AllocatedQty = new(int)
		}

		result := *existing.AllocatedQty + *orderdetails.AllocatedQty

		fmt.Printf("EXISTING ALLOC: %d, NEW ALLOC: %d, RESULT: %d\n",
			*existing.AllocatedQty,
			*orderdetails.AllocatedQty,
			result)

		*orderdetails.AllocatedQty = result
	}

	if err := services.DbUpdate(tx, &orderdetails, conditions); err != nil {
		return errors.New("failed updating requisition details")
	}
	if err := tx.Exec("EXEC sp_SetOrderStatus ?, ?", orderdetails.OrderDetailsID, orderType).Error; err != nil {
		return errors.New("failed executing stored procedure")
	}
	// orderdetailsat := models.OrderDetailsAt{
	// 	RefId:               orderdetails.Based_ID,
	// 	OrderDetailsContent: orderdetails.OrderDetailsContent,
	// 	At:                  at,
	// }

	// if err := services.DbInsert(tx, &orderdetailsat); err != nil {
	// 	return errors.New("failed creating order details at")
	// }

	return nil
}

func DeleteOrderDetail(tx *gorm.DB, orderdetails models.OrderDetails, at models.At, conditions map[string]interface{}) error {
	if err := services.DbDelete(tx, &orderdetails, conditions); err != nil {
		return errors.New("failed deleting order details")
	}

	orderdetailsat := models.OrderDetailsAt{
		RefId:               orderdetails.OrderDetailsID,
		OrderDetailsContent: orderdetails.OrderDetailsContent,
		At:                  at,
	}
	if err := services.DbInsert(tx, &orderdetailsat); err != nil {
		return errors.New("failed deleting order detail at")
	}

	return nil
}
