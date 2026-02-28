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

type SalesCalendarScheduleService struct{}

func NewSalesCalendarScheduleService() *SalesCalendarScheduleService {
	return &SalesCalendarScheduleService{}
}

// GET all sales schedules
func (s *SalesCalendarScheduleService) GetSalesSchedules(conditions map[string]interface{}) (*[]dispatching_models.SalesCalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	var schedules = &[]dispatching_models.SalesCalendarScheduleModel{}

	if tx.Error != nil {
		return schedules, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Find(schedules).Error; err != nil {
		fmt.Println("ERROR:", err)
		return schedules, fiber.StatusNotFound, errors.New("failed getting sales calendar schedules")
	}

	return schedules, fiber.StatusOK, nil
}

// GET a single sales schedule by ID
func (s *SalesCalendarScheduleService) GetSalesSchedule(conditions map[string]interface{}) (*dispatching_models.SalesCalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	var schedule = &dispatching_models.SalesCalendarScheduleModel{}

	if tx.Error != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).First(schedule).Error; err != nil {
		return schedule, fiber.StatusNotFound, errors.New("sales calendar schedule not found")
	}

	return schedule, fiber.StatusOK, nil
}

// CREATE a sales schedule
func (s *SalesCalendarScheduleService) CreateSalesSchedule(schedule *dispatching_models.SalesCalendarScheduleModel, at models.At) (*dispatching_models.SalesCalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &schedule); err != nil {

		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating sales schedule")
		}
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, err
	}

	atdata := dispatching_models.SalesCalendarScheduleModelAt{CalendarSchedulesBaseAt: dispatching_models.CalendarSchedulesBaseAt{
		RefId: schedule.ID,
		At:    at,
	}}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed creating sales schedule audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return schedule, fiber.StatusOK, nil
}

// UPDATE a sales schedule
func (s *SalesCalendarScheduleService) UpdateSalesSchedule(schedule *dispatching_models.SalesCalendarScheduleModel, conditions map[string]interface{}, at models.At) (*dispatching_models.SalesCalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &schedule, conditions); err != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed updating sales schedule")
	}

	atdata := dispatching_models.SalesCalendarScheduleModelAt{CalendarSchedulesBaseAt: dispatching_models.CalendarSchedulesBaseAt{
		RefId: schedule.ID,
		At:    at,
	}}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed creating sales schedule audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return schedule, fiber.StatusOK, nil
}

// DELETE a sales schedule
func (s *SalesCalendarScheduleService) DeleteSalesSchedule(conditions map[string]interface{}, at models.At) (*dispatching_models.SalesCalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return &dispatching_models.SalesCalendarScheduleModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	schedule, status, err := s.GetSalesSchedule(conditions)
	if err != nil {
		return schedule, status, errors.New("sales calendar schedule not found")
	}

	if err := services.DbDelete(tx, &schedule, conditions); err != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed deleting sales calendar schedule")
	}

	atdata := dispatching_models.SalesCalendarScheduleModelAt{CalendarSchedulesBaseAt: dispatching_models.CalendarSchedulesBaseAt{
		RefId: schedule.ID,
		At:    at,
	}}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed creating sales schedule audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return schedule, fiber.StatusOK, nil
}
