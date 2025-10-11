package dispatching_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

type CalendarEventService struct{}

func NewCalendarEventService() *CalendarEventService {
	return &CalendarEventService{}
}

func (s *CalendarEventService) GetCalendarEventsService(conditions map[string]interface{}) (*[]models.CalendarEventModel, int, error) {
	tx := initializers.DB.Begin()
	var events = &[]models.CalendarEventModel{}

	if tx.Error != nil {
		return events, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Find(events).Error; err != nil {
		fmt.Println("ERROR:", err)
		return events, 404, errors.New("failed getting calendar events")
	}

	return events, 200, nil
}

func (s *CalendarEventService) GetCalendarEventService(conditions map[string]interface{}) (*models.CalendarEventModel, int, error) {
	tx := initializers.DB.Begin()
	var event = &models.CalendarEventModel{}

	if tx.Error != nil {
		return event, 500, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).First(event).Error; err != nil {
		return event, 404, errors.New("calendar event not found")
	}

	return event, 200, nil
}

func (s *CalendarEventService) CreateCalendarEventService(event *models.CalendarEventModel, at models.At) (*models.CalendarEventModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return event, 500, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &event); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating calendar event")
		}
		tx.Rollback()
		return event, 500, err
	}

	atdata := models.PositionAt{RefId: event.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return event, 500, errors.New("failed creating event audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return event, 500, errors.New("failed to commit transaction")
	}

	return event, 201, nil
}

func (s *CalendarEventService) UpdateCalendarEventService(event *models.CalendarEventModel, conditions map[string]interface{}, at models.At) (*models.CalendarEventModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return event, 500, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &event, conditions); err != nil {
		return event, 500, errors.New("failed updating calendar event")
	}

	atdata := models.PositionAt{RefId: event.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return event, 500, errors.New("failed creating event audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return event, 500, errors.New("failed to commit transaction")
	}

	return event, 200, nil
}

func (s *CalendarEventService) DeleteCalendarEventService(conditions map[string]interface{}, at models.At) (*models.CalendarEventModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return &models.CalendarEventModel{}, 500, errors.New("failed to start DB transaction")
	}

	event, status, err := s.GetCalendarEventService(conditions)
	if err != nil {
		return event, status, errors.New("calendar event not found")
	}

	if err := services.DbDelete(tx, &event, conditions); err != nil {
		return event, 500, errors.New("failed deleting calendar event")
	}

	atdata := models.PositionAt{RefId: event.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return event, 500, errors.New("failed creating event audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return event, 500, errors.New("failed to commit transaction")
	}

	return event, 200, nil
}
