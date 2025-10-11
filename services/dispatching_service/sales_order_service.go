package dispatching_services

import (
	"errors"
	"net/http"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

type SalesOrderService struct{}

func NewSalesOrderService() *SalesOrderService {
	return &SalesOrderService{}
}

func (s *SalesOrderService) GetSalesOrdersService(conditions map[string]interface{}) ([]models.SalesOrderModel, int, error) {
	var orders = []models.SalesOrderModel{}
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return orders, 500, errors.New("failed to start DB transaction")
	}

	query := tx.Preload("DeliveryItems").Where(conditions).Find(&orders)
	if query.Error != nil {
		return nil, http.StatusInternalServerError, tx.Error
	}
	return orders, http.StatusOK, nil
}

func (s *SalesOrderService) GetSalesOrderService(conditions map[string]interface{}) (*models.SalesOrderModel, int, error) {
	var order = &models.SalesOrderModel{}
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return order, 500, errors.New("failed to start DB transaction")
	}

	query := tx.Preload("DeliveryItems").Where(conditions).First(&order)
	if query.Error != nil {
		return nil, http.StatusNotFound, tx.Error
	}
	return order, http.StatusOK, nil
}

func (s *SalesOrderService) CreateSalesOrderService(order *models.SalesOrderModel) (*models.SalesOrderModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return order, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Create(order).Error; err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return order, http.StatusCreated, nil
}

func (s *SalesOrderService) UpdateSalesOrderService(order *models.SalesOrderModel, conditions map[string]interface{}) (*models.SalesOrderModel, int, error) {
	var existing = &models.SalesOrderModel{}

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return order, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).First(&existing).Error; err != nil {
		return nil, http.StatusNotFound, err
	}
	if err := tx.Model(&existing).Updates(order).Error; err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return existing, http.StatusOK, nil
}

func (s *SalesOrderService) DeleteSalesOrderService(conditions map[string]interface{}) (bool, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return false, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Delete(&models.SalesOrderModel{}).Error; err != nil {
		return false, http.StatusInternalServerError, err
	}
	return true, http.StatusOK, nil
}
