package dispatching_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

type CalendarScheduleService struct{}

func NewCalendarScheduleService() *CalendarScheduleService {
	return &CalendarScheduleService{}
}

func (s *CalendarScheduleService) GetCalendarSchedulesService(conditions map[string]interface{}) (*[]models.CalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	var schedules = &[]models.CalendarScheduleModel{}

	if tx.Error != nil {
		return schedules, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Find(schedules).Error; err != nil {
		fmt.Println("ERROR:", err)
		return schedules, fiber.StatusInternalServerError, errors.New("failed getting calendar schedules")
	}

	return schedules, fiber.StatusOK, nil
}

func (s *CalendarScheduleService) GetCalendarScheduleService(conditions map[string]interface{}) (*models.CalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	var schedule = &models.CalendarScheduleModel{}

	if tx.Error != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).First(schedule).Error; err != nil {
		return schedule, fiber.StatusNotFound, errors.New("calendar schedule not found")
	}

	return schedule, fiber.StatusOK, nil
}

func (s *CalendarScheduleService) CreateCalendarScheduleService(schedule *models.CalendarScheduleModel, at models.At) (*models.CalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &schedule); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating schedule")
		}
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, err
	}

	atdata := models.CalendarScheduleAt{RefId: schedule.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed creating scheduleat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return schedule, fiber.StatusOK, nil
}

func (s *CalendarScheduleService) UpdateCalendarScheduleService(schedule *models.CalendarScheduleModel, conditions map[string]interface{}, at models.At) (*models.CalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &schedule, conditions); err != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed updating schedule")
	}

	atdata := models.CalendarScheduleAt{RefId: schedule.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed creating scheduleat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return schedule, fiber.StatusOK, nil
}

func (s *CalendarScheduleService) DeleteCalendarScheduleService(conditions map[string]interface{}, at models.At) (*models.CalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return &models.CalendarScheduleModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	schedule, status, err := s.GetCalendarScheduleService(conditions)
	if err != nil {
		return schedule, status, errors.New("calendar schedule not found")
	}

	if err := services.DbDelete(tx, &schedule, conditions); err != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed deleting calendar schedule")
	}

	atdata := models.CalendarScheduleAt{RefId: schedule.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed creating schedule audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return schedule, fiber.StatusOK, nil
}
