package dispatching_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type SalesOrderService struct{}

func NewSalesOrderService() *SalesOrderService {
	return &SalesOrderService{}
}

func (s *SalesOrderService) GetSalesOrdersService(conditions map[string]interface{}) ([]models.Order, int, error) {
	var orders = []models.Order{}
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return orders, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	// Read-only, and it was never committed or rolled back: every call left its
	// transaction - and the pooled connection under it - open.
	defer tx.Rollback()

	query := tx.Preload("Items").Omit("Order").
		Preload("Items.Item").
		// Preload("Items.Releases").Omit("Order").Omit("Item").Omit("Releases").
		// Preload("Items.Releases.Vehicle").
		// Preload("Items.Releases.DeliveryReceipt.TripCost").
		// Preload("DeliveryReceipts").Omit("Order").Omit("ItemReleases").
		// Preload("DeliveryReceipts.TripCost").Omit("DeliveryReceipt").
		// Preload("DeliveryReceipts.AttachedFile").Omit("DeliveryReceipt").
		Where(conditions).Find(&orders)
	if query.Error != nil {
		return nil, fiber.StatusNotFound, tx.Error
	}

	// for i := range orders {
	// 	for j := range orders[i].Items {
	// 		for k := range orders[i].Items[j].Releases {
	// 			orders[i].Items[j].Releases[k].Order = nil
	// 		}
	// 	}
	// }
	fillCustomerNames(tx, orders)
	return orders, fiber.StatusOK, nil
}

// The dispatching Sales Order screen showed CUSTOMER blank on every order (user-reported
// 2026-09-29): it reads customer_name, a denormalised copy the sales order form never
// fills in. BPI is the authority, resolved customer_id -> bpi.name as the logistics
// screens and resolveCustomerName already do; a stored name, when there is one, is kept.
// One query for the whole list rather than one per order.
func fillCustomerNames(tx *gorm.DB, orders []models.Order) {
	ids := []uint{}
	for _, o := range orders {
		if strings.TrimSpace(o.CustomerName) == "" && o.Customer_ID != 0 {
			ids = append(ids, o.Customer_ID)
		}
	}
	if len(ids) == 0 {
		return
	}

	var partners []models.Bpi
	if err := tx.Where("id IN ?", ids).Find(&partners).Error; err != nil {
		return // best effort: the list still loads, as it did before
	}
	names := map[uint]string{}
	for _, b := range partners {
		names[b.ID] = strings.TrimSpace(b.Name)
	}
	for i := range orders {
		if strings.TrimSpace(orders[i].CustomerName) == "" {
			orders[i].CustomerName = names[orders[i].Customer_ID]
		}
	}
}

func (s *SalesOrderService) GetSalesOrderService(conditions map[string]interface{}) (*models.Order, int, error) {
	var order = &models.Order{}
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return order, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}
	defer tx.Rollback()

	query := tx.Preload("Items").Omit("Order").
		Preload("Items.Item").
		// Preload("Items.Releases").Omit("Order").Omit("Item").Omit("Releases").
		// Preload("Items.Releases.Vehicle").
		// Preload("Items.Releases.DeliveryReceipt.TripCost").
		// Preload("DeliveryReceipts").Omit("Order").Omit("ItemReleases").
		// Preload("DeliveryReceipts.TripCost").Omit("DeliveryReceipt").
		// Preload("DeliveryReceipts.AttachedFile").Omit("DeliveryReceipt").
		Where(conditions).First(&order)

	if query.Error != nil {
		return nil, fiber.StatusNotFound, tx.Error
	}

	// for i := range order.Items {
	// 	for k := range order.Items[i].Releases {
	// 		order.Items[i].Releases[k].Order = nil
	// 	}
	// }

	order.CustomerName = resolveCustomerName(tx, order.CustomerName, order.Customer_ID)
	return order, fiber.StatusOK, nil
}

func (s *SalesOrderService) CreateSalesOrderService(order *models.Order, at models.At) (*models.Order, int, error) {
	tx := initializers.DB.Begin()

	if err := services.DbInsert(tx, &order); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating order")
		}
		tx.Rollback()
		return order, fiber.StatusInternalServerError, err
	}

	atdata := models.OrderAt{RefId: order.Order_ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return order, fiber.StatusInternalServerError, errors.New("failed creating orderat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return order, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return order, fiber.StatusOK, nil
}

func (s *SalesOrderService) UpdateSalesOrderService(order *models.Order, conditions map[string]interface{}, at models.At) (*models.Order, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return order, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &order, conditions); err != nil {
		return order, fiber.StatusInternalServerError, errors.New("failed updating order")
	}

	atdata := models.OrderAt{RefId: order.Order_ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return order, fiber.StatusInternalServerError, errors.New("failed creating orderat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return order, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return order, fiber.StatusOK, nil
}

// func (s *SalesOrderService) DeleteSalesOrderService(conditions map[string]interface{}, at models.At) (*models.Order, int, error) {
// 	tx := initializers.DB.Begin()
// 	if tx.Error != nil {
// 		return &models.Order{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
// 	}

// 	order, status, err := s.GetSalesOrderService(conditions)
// 	if err != nil {
// 		return order, status, errors.New("calendar order not found")
// 	}

// 	if err := services.DbDelete(tx, &order, conditions); err != nil {
// 		return order, fiber.StatusInternalServerError, errors.New("failed deleting calendar order")
// 	}

// 	atdata := models.OrderAt{RefId: order.Order_ID, At: at}
// 	if err := services.DbInsert(tx, &atdata); err != nil {
// 		tx.Rollback()
// 		return order, fiber.StatusInternalServerError, errors.New("failed creating order audit")
// 	}

// 	if err := tx.Commit().Error; err != nil {
// 		tx.Rollback()
// 		return order, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
// 	}

// 	return order, fiber.StatusOK, nil
// }
