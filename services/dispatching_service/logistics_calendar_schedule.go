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

type LogisticsCalendarScheduleService struct{}

func NewLogisticsCalendarScheduleService() *LogisticsCalendarScheduleService {
	return &LogisticsCalendarScheduleService{}
}

func (s *LogisticsCalendarScheduleService) GetLogisticsSchedules(conditions map[string]interface{}) (*[]dispatching_models.LogisticsCalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	var schedules = &[]dispatching_models.LogisticsCalendarScheduleModel{}

	if tx.Error != nil {
		return schedules, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Find(schedules).Error; err != nil {
		fmt.Println("ERROR:", err)
		return schedules, fiber.StatusNotFound, errors.New("failed getting logistics calendar schedules")
	}

	return schedules, fiber.StatusOK, nil
}

// GET a single logistics schedule by ID
func (s *LogisticsCalendarScheduleService) GetLogisticsSchedule(conditions map[string]interface{}) (*dispatching_models.LogisticsCalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	var schedule = &dispatching_models.LogisticsCalendarScheduleModel{}

	if tx.Error != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).First(schedule).Error; err != nil {
		return schedule, fiber.StatusNotFound, errors.New("logistics calendar schedule not found")
	}

	return schedule, fiber.StatusOK, nil
}

// CREATE a logistics schedule
func (s *LogisticsCalendarScheduleService) CreateLogisticsSchedule(schedule *dispatching_models.LogisticsCalendarScheduleModel, at models.At) (*dispatching_models.LogisticsCalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &schedule); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating logistics schedule")
		}
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, err
	}

	atdata := dispatching_models.LogisticsCalendarScheduleModelAt{CalendarSchedulesBaseAt: dispatching_models.CalendarSchedulesBaseAt{
		RefId: schedule.ID,
		At:    at,
	}}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed creating logistics schedule audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return schedule, fiber.StatusOK, nil
}

// UPDATE a logistics schedule
func (s *LogisticsCalendarScheduleService) UpdateLogisticsSchedule(schedule *dispatching_models.LogisticsCalendarScheduleModel, conditions map[string]interface{}, at models.At) (*dispatching_models.LogisticsCalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &schedule, conditions); err != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed updating logistics schedule")
	}

	atdata := dispatching_models.LogisticsCalendarScheduleModelAt{CalendarSchedulesBaseAt: dispatching_models.CalendarSchedulesBaseAt{
		RefId: schedule.ID,
		At:    at,
	}}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed creating logistics schedule audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return schedule, fiber.StatusOK, nil
}

// DELETE a logistics schedule
func (s *LogisticsCalendarScheduleService) DeleteLogisticsSchedule(conditions map[string]interface{}, at models.At) (*dispatching_models.LogisticsCalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return &dispatching_models.LogisticsCalendarScheduleModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	schedule, status, err := s.GetLogisticsSchedule(conditions)
	if err != nil {
		return schedule, status, errors.New("logistics calendar schedule not found")
	}

	if err := services.DbDelete(tx, &schedule, conditions); err != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed deleting logistics calendar schedule")
	}

	atdata := dispatching_models.LogisticsCalendarScheduleModelAt{CalendarSchedulesBaseAt: dispatching_models.CalendarSchedulesBaseAt{
		RefId: schedule.ID,
		At:    at,
	}}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed creating logistics schedule audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return schedule, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return schedule, fiber.StatusOK, nil
}
