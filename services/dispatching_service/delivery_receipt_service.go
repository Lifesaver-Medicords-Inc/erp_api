package dispatching_services

import (
	"errors"
	"net/http"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"gorm.io/gorm"
)

type DeliveryReceiptService struct{}

func NewDeliveryReceiptService() *DeliveryReceiptService {
	return &DeliveryReceiptService{}
}

func (s *DeliveryReceiptService) GetDeliveryReceiptsService(filters map[string]interface{}) ([]models.DeliveryReceiptModel, int, error) {

	var receipts = []models.DeliveryReceiptModel{}
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return receipts, 500, errors.New("failed to start DB transaction")
	}

	query := tx.Preload("SalesOrder").Preload("ReleasedItems")

	for key, val := range filters {
		query = query.Where(key+" = ?", val)
	}

	if err := query.Find(receipts).Error; err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return receipts, http.StatusOK, nil
}

func (s *DeliveryReceiptService) GetDeliveryReceiptService(filters map[string]interface{}) (*models.DeliveryReceiptModel, int, error) {

	var receipt = &models.DeliveryReceiptModel{}

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return receipt, 500, errors.New("failed to start DB transaction")
	}

	query := tx.Preload("SalesOrder").Preload("ReleasedItems")

	for key, val := range filters {
		query = query.Where(key+" = ?", val)
	}

	if err := query.First(receipt).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, http.StatusNotFound, err
		}
		return nil, http.StatusInternalServerError, err
	}

	return receipt, http.StatusOK, nil
}

func (s *DeliveryReceiptService) CreateDeliveryReceiptService(data *models.DeliveryReceiptModel) (*models.DeliveryReceiptModel, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return data, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Create(&data).Error; err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return data, http.StatusCreated, nil
}

func (s *DeliveryReceiptService) UpdateDeliveryReceiptService(id uint, update *models.DeliveryReceiptModel) (*models.DeliveryReceiptModel, int, error) {
	var receipt = &models.DeliveryReceiptModel{}

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return receipt, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.First(&receipt, id).Error; err != nil {
		return nil, http.StatusNotFound, err
	}

	if err := tx.Model(&receipt).Updates(update).Error; err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return receipt, http.StatusOK, nil
}

func (s *DeliveryReceiptService) DeleteDeliveryReceiptService(id uint) (bool, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return false, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Delete(&models.DeliveryReceiptModel{}, id).Error; err != nil {
		return false, http.StatusInternalServerError, err
	}
	return true, http.StatusOK, nil
}
