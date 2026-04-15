package journal_entry_services

import (
	// "errors"

	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type JournalEntryService struct{}

func NewJournalEntryService2() *JournalEntryService {
	return &JournalEntryService{}
}

func (s *JournalEntryService) GetCompanySetup(conditions map[string]interface{}) (interface{}, int, error) {
	var response models.CompanyCacheModel

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting smpc company setup")
	}

	return response, fiber.StatusOK, nil
}

func (s *JournalEntryService) GetCurrentJournal(conditions map[string]interface{}) (interface{}, int, error) {
	var response accounting_models.JournalEntryCurrent

	if err := services.DbGet(&response, conditions); err != nil {
		fmt.Println("DB ERROR:", err)
		return response, fiber.StatusInternalServerError, errors.New("failed getting journal entry")
	}

	// Define layout based on your format
	layout := "02/01/2006 3:04:05 pm"

	// Parse period_from
	periodFrom, err := time.Parse(layout, response.PeriodFrom)
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("invalid period_from format")
	}

	// Parse period_to
	periodTo, err := time.Parse(layout, response.PeriodTo)
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("invalid period_to format")
	}

	now := time.Now()

	// Check if current date is within the period
	if now.Before(periodFrom) || now.After(periodTo) {
		return nil, fiber.StatusNotFound, errors.New("no active journal entry for current date")
	}

	return response, fiber.StatusOK, nil
}

func (s *JournalEntryService) GetJournalEntry(conditions map[string]interface{}) (interface{}, int, error) {
	var response accounting_models.JournalEntryGet

	if err := services.DbGet(&response.JournalEntry, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting journal entry")
	}

	if err := services.DbGet(&response.JournalEntryDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting journal entry details")
	}

	return response, fiber.StatusOK, nil
}

func (s *JournalEntryService) CreateJournalEntry(body *accounting_models.JournalEntryBody, at models.At) (*accounting_models.JournalEntryBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	// Rollback once, automatically, unless committed
	defer tx.Rollback()

	nextDocNo, err := utils.NextDocNo(tx, new(accounting_models.JournalEntry), "doc_no")
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed getting next doc number")
	}

	body.JournalEntry.DocNo = nextDocNo

	// Insert main Journal Entry
	if err := services.DbInsert(tx, &body.JournalEntry); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return body, fiber.StatusConflict, errors.New("duplicate record error")
		}
		return body, fiber.StatusInternalServerError, errors.New("failed creating journal entry")
	}

	// Insert Journal Entry Details
	if err := s.CreateJournalEntryDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Insert audit record for the main request
	atdata := accounting_models.JournalEntryAt{
		RefId:               body.JournalEntry.ID,
		JournalEntryContent: body.JournalEntry.JournalEntryContent,
		At:                  at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating journal entry at")
	}

	// Commit once
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return body, fiber.StatusOK, nil
}

func (s *JournalEntryService) CreateJournalEntryDetails(tx *gorm.DB, body *accounting_models.JournalEntryBody, at models.At) error {
	for i := range body.JournalEntryDetails {
		detail := &body.JournalEntryDetails[i]
		detail.JournalEntryId = body.JournalEntry.ID // assign FK to parent

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating journal entry details")
		}

		// Audit trail for each detail
		atdataDetail := accounting_models.JournalEntryDetailsAt{
			RefId:                      detail.ID,
			JournalEntryDetailsContent: detail.JournalEntryDetailsContent,
			At:                         at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating journal entry details at")
		}
	}
	return nil
}

func (s *JournalEntryService) UpdateJournalEntry(body *accounting_models.JournalEntryBody, conditions map[string]interface{}, at models.At) (*accounting_models.JournalEntryBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	// Rollback once, automatically, unless committed
	defer tx.Rollback()

	// Update main Journal Entry
	if err := services.DbUpdate(tx, &body.JournalEntry, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating journal entry")
	}

	// Handle details
	if err := s.UpdateJournalEntryDetails(tx, body, conditions, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Audit record for main request
	atdata := accounting_models.JournalEntryAt{
		RefId:               body.JournalEntry.ID,
		JournalEntryContent: body.JournalEntry.JournalEntryContent,
		At:                  at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating journal entry at")
	}

	// Commit once
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return body, fiber.StatusOK, nil
}

func (s *JournalEntryService) UpdateJournalEntryDetails(tx *gorm.DB, body *accounting_models.JournalEntryBody, conditions map[string]interface{}, at models.At) error {
	for i := range body.JournalEntryDetails {
		detail := &body.JournalEntryDetails[i]
		detail.JournalEntryId = body.JournalEntry.ID // ensure FK is set

		if detail.ID == 0 {
			if err := services.DbInsert(tx, detail); err != nil {
				return errors.New("failed creating journal entry details")
			}
		} else {
			if err := services.DbUpdate(tx, detail, conditions); err != nil {
				return errors.New("failed updating journal entry details")
			}
		}

		// Audit record for each detail
		atdataDetail := accounting_models.JournalEntryDetailsAt{
			RefId:                      detail.ID,
			JournalEntryDetailsContent: detail.JournalEntryDetailsContent,
			At:                         at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating journal entry details at")
		}
	}
	return nil
}

func (s *JournalEntryService) DeleteJournalEntry(body *accounting_models.JournalEntryBody, at models.At) (*accounting_models.JournalEntryBody, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	// Rollback once, automatically, unless committed
	defer tx.Rollback()

	// Delete main Journal Entry
	if err := services.DbDelete(tx, &body.JournalEntry, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting journal entry")
	}

	if err := s.DeleteJournalEntryDetails(tx, body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Audit record for main request
	atdata := accounting_models.JournalEntryAt{RefId: body.JournalEntry.ID, JournalEntryContent: body.JournalEntry.JournalEntryContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating journal entry at")
	}

	// Commit once
	if err := tx.Commit().Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed committing transaction")
	}

	return body, fiber.StatusOK, nil
}

func (s *JournalEntryService) DeleteJournalEntryDetails(tx *gorm.DB, body *accounting_models.JournalEntryBody, at models.At) error {
	// Delete all details
	if err := services.DbDelete(tx, &accounting_models.JournalEntryDetails{}, map[string]interface{}{"journal_entry_id": body.JournalEntry.ID}); err != nil {
		return errors.New("failed deleting all journal entry details")
	}

	// Optionally fetch deleted details for audit trail
	var deletedDetails []accounting_models.JournalEntryDetails
	if err := tx.Unscoped().Where("journal_entry_id = ?", body.JournalEntry.ID).Find(&deletedDetails).Error; err == nil {
		for _, detail := range deletedDetails {
			atdataDetail := accounting_models.JournalEntryDetailsAt{
				RefId:                      detail.ID,
				JournalEntryDetailsContent: detail.JournalEntryDetailsContent,
				At:                         at,
			}
			if err := services.DbInsert(tx, &atdataDetail); err != nil {
				return errors.New("failed creating journal entry details audit record")
			}
		}
	}
	return nil
}

func (s *JournalEntryService) AutoInsertJournalEntry(body *accounting_models.JournalEntryDetails, date string, at models.At) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return errors.New("failed to start DB transaction")
	}

	// Rollback once, automatically, unless committed
	defer tx.Rollback()

	// Insert main Journal Entry Details
	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("duplicate record error")
		}
		return errors.New("failed creating journal entry details auto insert")
	}

	// Insert audit record for the main request
	atdata := accounting_models.JournalEntryDetailsAt{
		RefId:                      body.ID,
		JournalEntryDetailsContent: body.JournalEntryDetailsContent,
		At:                         at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return errors.New("failed creating journal entry details auto insert at")
	}

	// Commit once
	if err := tx.Commit().Error; err != nil {
		return errors.New("failed committing transaction")
	}

	return nil
}
