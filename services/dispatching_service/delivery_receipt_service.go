package dispatching_services

import (
	"errors"
	"strings"
	"time"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type DeliveryReceiptService struct {
	CalendarScheduleService *CalendarScheduleService
}

func NewDeliveryReceiptService(calendarScheduleService *CalendarScheduleService) *DeliveryReceiptService {
	return &DeliveryReceiptService{
		CalendarScheduleService: calendarScheduleService,
	}
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
		return nil, 500, err
	}
	return receipts, 200, nil
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
			return nil, 404, err
		}
		return nil, 500, err
	}

	return receipt, 200, nil
}

func (s *DeliveryReceiptService) CreateDeliveryReceiptService(data *models.DeliveryReceiptModel, at models.At) (*models.DeliveryReceiptModel, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return data, 500, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &data); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {

			err = errors.New("duplicate record error")
		} else {

			err = errors.New("failed creating delivery receipt")
		}
		tx.Rollback()
		return data, 500, err
	}

	atdata := models.CalendarScheduleAt{RefId: data.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return data, 500, errors.New("failed creating receiptat")
	}

	// ✅ Create logistics calendar schedule
	schedule := models.CalendarScheduleModel{
		RelatedOrderID: &data.OrderID,
		CalendarScheduleContent: models.CalendarScheduleContent{
			DepartmentType: "Logistics",
			Title:          "",
			Description:    "",
			StartDate:      data.DeliveryDate,
			EndDate:        data.DeliveryDate.Add(2 * time.Hour),
			ReferenceType:  "DeliveryReceipt",
			ReferenceId:    data.ID,
		},
	}

	if _, _, err := s.CalendarScheduleService.CreateCalendarScheduleService(&schedule, at); err != nil {
		tx.Rollback()
		return data, 500, errors.New("failed creating calendar schedule")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return data, 500, errors.New("failed to commit transaction")
	}

	return data, 201, nil
}

func (s *DeliveryReceiptService) UpdateDeliveryReceiptService(update *models.DeliveryReceiptModel, conditions map[string]interface{}, at models.At) (*models.DeliveryReceiptModel, int, error) {
	var receipt = &models.DeliveryReceiptModel{}

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return receipt, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.First(&receipt, conditions).Error; err != nil {
		return nil, 404, err
	}

	if err := services.DbUpdate(tx, &receipt, conditions); err != nil {
		return receipt, 500, errors.New("failed updating receipt")
	}

	atdata := models.DeliveryReceiptAt{RefId: receipt.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return receipt, 500, errors.New("failed creating receiptat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return receipt, 500, errors.New("failed to commit transaction")
	}
	return receipt, 200, nil
}

func (s *DeliveryReceiptService) DeleteDeliveryReceiptService(conditions map[string]interface{}, at models.At) (*models.DeliveryReceiptModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.DeliveryReceiptModel{}, 500, errors.New("failed to start DB transaction")
	}
	release, status, err := s.GetDeliveryReceiptService(conditions)
	if err != nil {
		return release, status, errors.New("delivery receipt not found")
	}

	if err := services.DbDelete(tx, &release, conditions); err != nil {
		return release, 500, errors.New("failed deleting delivery receipt")
	}

	atdata := models.DeliveryReceiptAt{RefId: release.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return release, 500, errors.New("failed creating receipt audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return release, 500, errors.New("failed to commit transaction")
	}

	return release, 200, nil
}
