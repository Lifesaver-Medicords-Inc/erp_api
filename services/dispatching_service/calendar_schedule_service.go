package dispatching_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type CalendarScheduleService struct{}

func NewCalendarScheduleService() *CalendarScheduleService {
	return &CalendarScheduleService{}
}

// Reads don't need a transaction — querying initializers.DB directly avoids
// opening a tx that never gets committed/rolled back (a connection leak).
func (s *CalendarScheduleService) GetCalendarSchedulesService(conditions map[string]interface{}) (*[]models.CalendarScheduleModel, int, error) {
	var schedules = &[]models.CalendarScheduleModel{}

	if err := initializers.DB.Where(conditions).Find(schedules).Error; err != nil {
		fmt.Println("ERROR:", err)
		return schedules, fiber.StatusInternalServerError, errors.New("failed getting calendar schedules")
	}

	return schedules, fiber.StatusOK, nil
}

func (s *CalendarScheduleService) GetCalendarScheduleService(conditions map[string]interface{}) (*models.CalendarScheduleModel, int, error) {
	var schedule = &models.CalendarScheduleModel{}

	if err := initializers.DB.Where(conditions).First(schedule).Error; err != nil {
		return schedule, fiber.StatusNotFound, errors.New("calendar schedule not found")
	}

	return schedule, fiber.StatusOK, nil
}

// CreateCalendarScheduleService inserts within the caller's transaction so it
// commits/rolls back atomically with whatever operation is creating this schedule
// (e.g. a Delivery Receipt). It does not begin or commit a transaction itself.
func (s *CalendarScheduleService) CreateCalendarScheduleService(tx *gorm.DB, schedule *models.CalendarScheduleModel, at models.At) (*models.CalendarScheduleModel, int, error) {
	if err := services.DbInsert(tx, &schedule); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return schedule, fiber.StatusInternalServerError, errors.New("duplicate record error")
		}
		return schedule, fiber.StatusInternalServerError, errors.New("failed creating schedule")
	}

	atdata := models.CalendarScheduleAt{RefId: schedule.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed creating scheduleat")
	}

	return schedule, fiber.StatusOK, nil
}

// UpdateCalendarScheduleService updates within the caller's transaction so it
// commits/rolls back atomically with whatever operation is updating this schedule
// (e.g. a Delivery Receipt). It does not begin or commit a transaction itself.
func (s *CalendarScheduleService) UpdateCalendarScheduleService(tx *gorm.DB, schedule *models.CalendarScheduleModel, conditions map[string]interface{}, at models.At) (*models.CalendarScheduleModel, int, error) {
	if err := services.DbUpdate(tx, &schedule, conditions); err != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed updating schedule")
	}

	atdata := models.CalendarScheduleAt{RefId: schedule.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return schedule, fiber.StatusInternalServerError, errors.New("failed creating scheduleat")
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
		tx.Rollback()
		return schedule, status, errors.New("calendar schedule not found")
	}

	if err := services.DbDelete(tx, &schedule, conditions); err != nil {
		tx.Rollback()
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
