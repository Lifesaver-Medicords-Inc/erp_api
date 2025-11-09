package dispatching_services

import (
	"errors"
	"strings"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

type SalesOrderService struct{}

func NewSalesOrderService() *SalesOrderService {
	return &SalesOrderService{}
}

func (s *SalesOrderService) GetSalesOrdersService(conditions map[string]interface{}) ([]models.Order, int, error) {
	var orders = []models.Order{}
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return orders, 500, errors.New("failed to start DB transaction")
	}

	query := tx.Preload("Items").Omit("Order").
		Preload("Items.Item").
		Preload("Items.Releases").Omit("Order").Omit("Item").Omit("Releases").
		Preload("Items.Releases.Vehicle").
		Preload("Items.Releases.DeliveryReceipt.TripCost").
		Preload("DeliveryReceipts").Omit("Order").Omit("ItemReleases").
		Preload("DeliveryReceipts.TripCost").Omit("DeliveryReceipt").
		Preload("DeliveryReceipts.AttachedFile").Omit("DeliveryReceipt").
		Where(conditions).Find(&orders)
	if query.Error != nil {
		return nil, 404, tx.Error
	}

	for i := range orders {
		for j := range orders[i].Items {
			for k := range orders[i].Items[j].Releases {
				orders[i].Items[j].Releases[k].Order = nil
			}
		}
	}
	return orders, 200, nil
}

func (s *SalesOrderService) GetSalesOrderService(conditions map[string]interface{}) (*models.Order, int, error) {
	var order = &models.Order{}
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return order, 500, errors.New("failed to start DB transaction")
	}

	query := tx.Preload("Items").Omit("Order").
		Preload("Items.Item").
		Preload("Items.Releases").Omit("Order").Omit("Item").Omit("Releases").
		Preload("Items.Releases.Vehicle").
		Preload("Items.Releases.DeliveryReceipt.TripCost").
		Preload("DeliveryReceipts").Omit("Order").Omit("ItemReleases").
		Preload("DeliveryReceipts.TripCost").Omit("DeliveryReceipt").
		Preload("DeliveryReceipts.AttachedFile").Omit("DeliveryReceipt").
		Where(conditions).First(&order)

	if query.Error != nil {
		return nil, 404, tx.Error
	}

	for i := range order.Items {
		for k := range order.Items[i].Releases {
			order.Items[i].Releases[k].Order = nil
		}
	}

	return order, 200, nil
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
		return order, 500, err
	}

	atdata := models.OrderAt{RefId: order.Order_ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return order, 500, errors.New("failed creating orderat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return order, 500, errors.New("failed to commit transaction")
	}

	return order, 200, nil
}

func (s *SalesOrderService) UpdateSalesOrderService(order *models.Order, conditions map[string]interface{}, at models.At) (*models.Order, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return order, 500, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &order, conditions); err != nil {
		return order, 500, errors.New("failed updating order")
	}

	atdata := models.OrderAt{RefId: order.Order_ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return order, 500, errors.New("failed creating orderat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return order, 500, errors.New("failed to commit transaction")
	}

	return order, 200, nil
}

func (s *SalesOrderService) DeleteSalesOrderService(conditions map[string]interface{}, at models.At) (*models.Order, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return &models.Order{}, 500, errors.New("failed to start DB transaction")
	}

	order, status, err := s.GetSalesOrderService(conditions)
	if err != nil {
		return order, status, errors.New("calendar order not found")
	}

	if err := services.DbDelete(tx, &order, conditions); err != nil {
		return order, 500, errors.New("failed deleting calendar order")
	}

	atdata := models.OrderAt{RefId: order.Order_ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return order, 500, errors.New("failed creating order audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return order, 500, errors.New("failed to commit transaction")
	}

	return order, 200, nil
}
