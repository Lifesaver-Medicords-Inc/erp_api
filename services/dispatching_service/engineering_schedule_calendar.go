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

type EngineeringCalendarScheduleService struct{}

func NewEngineeringCalendarScheduleService() *EngineeringCalendarScheduleService {
	return &EngineeringCalendarScheduleService{}
}

// GET all engineering schedules
func (s *EngineeringCalendarScheduleService) GetEngineeringSchedules(conditions map[string]interface{}) (*[]dispatching_models.EngineeringCalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	var schedules = &[]dispatching_models.EngineeringCalendarScheduleModel{}

	if tx.Error != nil {
		return schedules, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Find(schedules).Error; err != nil {
		fmt.Println("ERROR:", err)
		return schedules, 404, errors.New("failed getting engineering calendar schedules")
	}

	return schedules, 200, nil
}

// GET a single engineering schedule by ID
func (s *EngineeringCalendarScheduleService) GetEngineeringSchedule(conditions map[string]interface{}) (*dispatching_models.EngineeringCalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	var schedule = &dispatching_models.EngineeringCalendarScheduleModel{}

	if tx.Error != nil {
		return schedule, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).First(schedule).Error; err != nil {
		return schedule, 404, errors.New("engineering calendar schedule not found")
	}

	return schedule, 200, nil
}

// CREATE an engineering schedule
func (s *EngineeringCalendarScheduleService) CreateEngineeringSchedule(schedule *dispatching_models.EngineeringCalendarScheduleModel, at models.At) (*dispatching_models.EngineeringCalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return schedule, 500, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &schedule); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating engineering schedule")
		}
		tx.Rollback()
		return schedule, 500, err
	}

	atdata := dispatching_models.EngineeringCalendarScheduleModelAt{CalendarSchedulesBaseAt: dispatching_models.CalendarSchedulesBaseAt{
		RefId: schedule.ID,
		At:    at,
	}}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return schedule, 500, errors.New("failed creating engineering schedule audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return schedule, 500, errors.New("failed to commit transaction")
	}

	return schedule, 200, nil
}

// UPDATE an engineering schedule
func (s *EngineeringCalendarScheduleService) UpdateEngineeringSchedule(schedule *dispatching_models.EngineeringCalendarScheduleModel, conditions map[string]interface{}, at models.At) (*dispatching_models.EngineeringCalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return schedule, 500, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &schedule, conditions); err != nil {
		return schedule, 500, errors.New("failed updating engineering schedule")
	}

	atdata := dispatching_models.EngineeringCalendarScheduleModelAt{CalendarSchedulesBaseAt: dispatching_models.CalendarSchedulesBaseAt{
		RefId: schedule.ID,
		At:    at,
	}}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return schedule, 500, errors.New("failed creating engineering schedule audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return schedule, 500, errors.New("failed to commit transaction")
	}

	return schedule, 200, nil
}

// DELETE an engineering schedule
func (s *EngineeringCalendarScheduleService) DeleteEngineeringSchedule(conditions map[string]interface{}, at models.At) (*dispatching_models.EngineeringCalendarScheduleModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return &dispatching_models.EngineeringCalendarScheduleModel{}, 500, errors.New("failed to start DB transaction")
	}

	schedule, status, err := s.GetEngineeringSchedule(conditions)
	if err != nil {
		return schedule, status, errors.New("engineering calendar schedule not found")
	}

	if err := services.DbDelete(tx, &schedule, conditions); err != nil {
		return schedule, 500, errors.New("failed deleting engineering calendar schedule")
	}

	atdata := dispatching_models.EngineeringCalendarScheduleModelAt{CalendarSchedulesBaseAt: dispatching_models.CalendarSchedulesBaseAt{
		RefId: schedule.ID,
		At:    at,
	}}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return schedule, 500, errors.New("failed creating engineering schedule audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return schedule, 500, errors.New("failed to commit transaction")
	}

	return schedule, 200, nil
}
