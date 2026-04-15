package dispatching_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"
	"github.com/pierceperado/smpc/services"
)

type CalendarCategoryService struct{}

func NewCalendarCategoryService() *CalendarCategoryService {
	return &CalendarCategoryService{}
}

func (s *CalendarCategoryService) GetCalendarCategoriesService(conditions map[string]interface{}) (*[]dispatching_models.CalendarCategoryModel, int, error) {
	tx := initializers.DB.Begin()
	var categories = &[]dispatching_models.CalendarCategoryModel{}

	if tx.Error != nil {
		return categories, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbGet(categories, conditions); err != nil {
		fmt.Println("ERROR:", err)
		return categories, fiber.StatusInternalServerError, errors.New("failed getting calendar categories")
	}

	return categories, fiber.StatusOK, nil
}

func (s *CalendarCategoryService) GetCalendarCategoryService(conditions map[string]interface{}) (*dispatching_models.CalendarCategoryModel, int, error) {
	tx := initializers.DB.Begin()
	var categories = &dispatching_models.CalendarCategoryModel{}

	if tx.Error != nil {
		return categories, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbGet(categories, conditions); err != nil {
		return categories, fiber.StatusNotFound, errors.New("calendar category not found")
	}

	return categories, fiber.StatusOK, nil
}

func (s *CalendarCategoryService) CreateCalendarCategoryService(category *dispatching_models.CalendarCategoryModel, at models.At) (*dispatching_models.CalendarCategoryModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return category, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &category); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating category")
		}
		tx.Rollback()
		return category, fiber.StatusInternalServerError, err
	}

	atdata := dispatching_models.CalendarCategoryAt{RefId: category.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return category, fiber.StatusInternalServerError, errors.New("failed creating categoryat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return category, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return category, fiber.StatusOK, nil
}

func (s *CalendarCategoryService) UpdateCalendarCategoryService(category *dispatching_models.CalendarCategoryModel, conditions map[string]interface{}, at models.At) (*dispatching_models.CalendarCategoryModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return category, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &category, conditions); err != nil {
		return category, fiber.StatusInternalServerError, errors.New("failed updating category")
	}

	atdata := dispatching_models.CalendarCategoryAt{RefId: category.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return category, fiber.StatusInternalServerError, errors.New("failed creating categoryat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return category, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return category, fiber.StatusOK, nil
}

func (s *CalendarCategoryService) DeleteCalendarCategoryService(conditions map[string]interface{}, at models.At) (*dispatching_models.CalendarCategoryModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return &dispatching_models.CalendarCategoryModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	category, status, err := s.GetCalendarCategoryService(conditions)
	if err != nil {
		return category, status, errors.New("calendar category not found")
	}

	if err := services.DbDelete(tx, &category, conditions); err != nil {
		return category, fiber.StatusInternalServerError, errors.New("failed deleting calendar category")
	}

	atdata := dispatching_models.CalendarCategoryAt{RefId: category.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return category, fiber.StatusInternalServerError, errors.New("failed creating categoryat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return category, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return category, fiber.StatusOK, nil
}
