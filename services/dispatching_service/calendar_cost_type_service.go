package dispatching_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"
	"github.com/pierceperado/smpc/services"
)

type CalendarCostTypeService struct{}

func NewCalendarCostTypeService() *CalendarCostTypeService {
	return &CalendarCostTypeService{}
}

func (s *CalendarCostTypeService) GetCalendarCostTypesService(conditions map[string]interface{}) (*[]dispatching_models.CalendarCostTypeModel, int, error) {
	tx := initializers.DB.Begin()
	var costType = &[]dispatching_models.CalendarCostTypeModel{}

	if tx.Error != nil {
		return costType, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Find(costType).Error; err != nil {
		fmt.Println("ERROR:", err)
		return costType, 404, errors.New("failed getting calendar cost types")
	}

	return costType, 200, nil
}

func (s *CalendarCostTypeService) GetCalendarCostTypeService(conditions map[string]interface{}) (*dispatching_models.CalendarCostTypeModel, int, error) {
	tx := initializers.DB.Begin()
	var costType = &dispatching_models.CalendarCostTypeModel{}

	if tx.Error != nil {
		return costType, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).First(costType).Error; err != nil {
		return costType, 404, errors.New("calendar cost type not found")
	}

	return costType, 200, nil
}

func (s *CalendarCostTypeService) CreateCalendarCostTypeService(costType *dispatching_models.CalendarCostTypeModel, at models.At) (*dispatching_models.CalendarCostTypeModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return costType, 500, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &costType); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {

			err = errors.New("duplicate record error")
		} else {

			err = errors.New("failed creating cost type")
		}
		tx.Rollback()
		return costType, 500, err
	}

	atdata := dispatching_models.CalendarCostTypeAt{RefId: costType.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return costType, 500, errors.New("failed creating costtypeat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return costType, 500, errors.New("failed to commit transaction")
	}

	return costType, 200, nil
}

func (s *CalendarCostTypeService) UpdateCalendarCostTypeService(costType *dispatching_models.CalendarCostTypeModel, conditions map[string]interface{}, at models.At) (*dispatching_models.CalendarCostTypeModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return costType, 500, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &costType, conditions); err != nil {
		return costType, 500, errors.New("failed updating cost type")
	}

	atdata := dispatching_models.CalendarCostTypeAt{RefId: costType.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return costType, 500, errors.New("failed creating  costtypeat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return costType, 500, errors.New("failed to commit transaction")
	}

	return costType, 200, nil
}

func (s *CalendarCostTypeService) DeleteCalendarCostTypeService(conditions map[string]interface{}, at models.At) (*dispatching_models.CalendarCostTypeModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return &dispatching_models.CalendarCostTypeModel{}, 500, errors.New("failed to start DB transaction")
	}

	costType, status, err := s.GetCalendarCostTypeService(conditions)
	if err != nil {
		return costType, status, errors.New("calendar cost type not found")
	}

	if err := services.DbDelete(tx, &costType, conditions); err != nil {
		return costType, 500, errors.New("failed deleting calendar cost type")
	}

	atdata := dispatching_models.CalendarCostTypeAt{RefId: costType.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return costType, 500, errors.New("failed creating costtype audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return costType, 500, errors.New("failed to commit transaction")
	}

	return costType, 200, nil
}
