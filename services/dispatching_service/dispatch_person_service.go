package dispatching_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// Roles a dispatch person can hold. §13.3 names exactly these two ("select driver,
// helper, and vehicle first") and §17 defines no list for them, so nothing else is
// accepted - CLAUDE.md forbids inventing dropdown values.
const (
	RoleDriver = "DRIVER"
	RoleHelper = "HELPER"
)

func IsValidDispatchRole(role string) bool {
	switch strings.ToUpper(strings.TrimSpace(role)) {
	case RoleDriver, RoleHelper:
		return true
	}
	return false
}

type DispatchPersonService struct{}

func NewDispatchPersonService() *DispatchPersonService {
	return &DispatchPersonService{}
}

// Listing is deliberately unfiltered by default so Setup can show inactive people
// (they are never deleted - see the model). The schedule pickers pass
// {"is_active": true} to get only the people who can still be assigned.
func (s *DispatchPersonService) GetDispatchPeople(conditions map[string]interface{}) ([]dispatching_models.DispatchPerson, int, error) {
	var people []dispatching_models.DispatchPerson

	if err := initializers.DB.Where(conditions).Order("full_name").Find(&people).Error; err != nil {
		return people, fiber.StatusInternalServerError, errors.New("failed getting dispatch people")
	}

	return people, fiber.StatusOK, nil
}

func (s *DispatchPersonService) GetDispatchPerson(conditions map[string]interface{}) (*dispatching_models.DispatchPerson, int, error) {
	var person = &dispatching_models.DispatchPerson{}

	if err := initializers.DB.Where(conditions).First(person).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return person, fiber.StatusNotFound, errors.New("dispatch person not found")
		}
		return person, fiber.StatusInternalServerError, errors.New("failed getting dispatch person")
	}

	return person, fiber.StatusOK, nil
}

func (s *DispatchPersonService) CreateDispatchPerson(data *dispatching_models.DispatchPerson, at models.At) (*dispatching_models.DispatchPerson, int, error) {
	if status, err := validateDispatchPerson(data); err != nil {
		return data, status, err
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return data, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	data.ID = 0 // never trust a client-supplied id

	if err := services.DbInsert(tx, data); err != nil {
		tx.Rollback()
		return data, fiber.StatusInternalServerError, errors.New("failed creating dispatch person")
	}

	atdata := dispatching_models.DispatchPersonAt{
		RefId:                 data.ID,
		DispatchPersonContent: data.DispatchPersonContent,
		At:                    at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return data, fiber.StatusInternalServerError, errors.New("failed creating dispatch person audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return data, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return data, fiber.StatusCreated, nil
}

func (s *DispatchPersonService) UpdateDispatchPerson(update *dispatching_models.DispatchPerson, conditions map[string]interface{}, at models.At) (*dispatching_models.DispatchPerson, int, error) {
	if status, err := validateDispatchPerson(update); err != nil {
		return update, status, err
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return update, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	var existing dispatching_models.DispatchPerson
	if err := tx.Where(conditions).First(&existing).Error; err != nil {
		tx.Rollback()
		return update, fiber.StatusNotFound, errors.New("dispatch person not found")
	}

	if err := services.DbUpdate(tx, update, conditions); err != nil {
		tx.Rollback()
		return update, fiber.StatusInternalServerError, errors.New("failed updating dispatch person")
	}

	// Renaming a person must not silently rewrite the trips they already went on -
	// tbl_dispatching_schedule_people carries its own copy of the name for exactly
	// that reason (see the model). Only the schedule's driver_name mirror is kept in
	// step, and only where this person is currently the assigned driver.
	if err := tx.Exec(`
		UPDATE sch
		SET sch.driver_name = ?
		FROM tbl_dispatching_logistics_calendar_schedule sch
			INNER JOIN tbl_dispatching_schedule_people sp ON sp.schedule_id = sch.id
		WHERE sp.person_id = ? AND sp.role = ?
	`, update.FullName, existing.ID, RoleDriver).Error; err != nil {
		tx.Rollback()
		return update, fiber.StatusInternalServerError, errors.New("failed syncing driver name on schedules")
	}

	atdata := dispatching_models.DispatchPersonAt{
		RefId:                 existing.ID,
		DispatchPersonContent: update.DispatchPersonContent,
		At:                    at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return update, fiber.StatusInternalServerError, errors.New("failed creating dispatch person audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return update, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return update, fiber.StatusOK, nil
}

// Deactivate rather than delete. A person who has left stops appearing in the schedule
// pickers, but every trip they were on keeps their name - the same rule §4.4.1 applies
// to an inactive warehouse ("all past data is retained").
func (s *DispatchPersonService) DeactivateDispatchPerson(conditions map[string]interface{}, at models.At) (*dispatching_models.DispatchPerson, int, error) {
	var person = &dispatching_models.DispatchPerson{}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return person, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).First(person).Error; err != nil {
		tx.Rollback()
		return person, fiber.StatusNotFound, errors.New("dispatch person not found")
	}

	person.IsActive = false

	if err := tx.Model(&dispatching_models.DispatchPerson{}).
		Where("id = ?", person.ID).
		Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return person, fiber.StatusInternalServerError, errors.New("failed deactivating dispatch person")
	}

	atdata := dispatching_models.DispatchPersonAt{
		RefId:                 person.ID,
		DispatchPersonContent: person.DispatchPersonContent,
		At:                    at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return person, fiber.StatusInternalServerError, errors.New("failed creating dispatch person audit")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return person, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return person, fiber.StatusOK, nil
}

func validateDispatchPerson(data *dispatching_models.DispatchPerson) (int, error) {
	if data == nil {
		return fiber.StatusBadRequest, errors.New("no dispatch person supplied")
	}

	data.FullName = strings.TrimSpace(data.FullName)
	if data.FullName == "" {
		return fiber.StatusBadRequest, errors.New("full name is required")
	}

	data.Role = strings.ToUpper(strings.TrimSpace(data.Role))
	if !IsValidDispatchRole(data.Role) {
		return fiber.StatusBadRequest, errors.New("role must be DRIVER or HELPER")
	}

	return fiber.StatusOK, nil
}
