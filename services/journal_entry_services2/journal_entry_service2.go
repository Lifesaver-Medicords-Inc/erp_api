package journal_entry1_services2

import (
	// "errors"

	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type JournalEntryBody struct {
	JournalEntry        accounting_models.JournalEntry2          `json:"journal_entry"`
	JournalEntryDetails []accounting_models.JournalEntryDetails2 `json:"journal_entry_details"`
}

type JournalEntryGet struct {
	JournalEntry        []accounting_models.JournalEntry2        `json:"journal_entry"`
	JournalEntryDetails []accounting_models.JournalEntryDetails2 `json:"journal_entry_details"`
}

func GetCompanySetup(conditions map[string]interface{}) (interface{}, int, error) {
	var response models.CompanyCacheModel

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting smpc company setup")
	}

	return response, 0, nil
}

func GetJournalEntry(conditions map[string]interface{}) (interface{}, int, error) {
	var response JournalEntryGet

	if err := services.DbGet(&response.JournalEntry, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting item request")
	}

	if err := services.DbGet(&response.JournalEntryDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New(" failed getting item request details")
	}

	return response, 0, nil
}

func CreateJournalEntry(c *fiber.Ctx, tx *gorm.DB) (interface{}, int, error) {
	var body JournalEntryBody

	//Parse the full request body (main + details)
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	//Insert main Journal Entry record
	if err := services.DbInsert(tx, &body.JournalEntry); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating journal entry")
	}

	generatedDocNo := utils.DocNoGenerator(body.JournalEntry.ID)
	body.JournalEntry.DocNo = generatedDocNo

	if err := tx.Model(&body.JournalEntry).Update("doc_no", body.JournalEntry.DocNo).Error; err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating journal entry doc")
	}

	//Prepare the "at" data
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	// Insert Journal Entry Details
	if err := CreateJournalEntryDetails(tx, &body, at); err != nil {
		return body.JournalEntry, fiber.StatusInternalServerError, err
	}

	//Insert audit record for the main request
	atdata := accounting_models.JournalEntry2At{
		RefId:                body.JournalEntry.ID,
		JournalEntry2Content: body.JournalEntry.JournalEntry2Content,
		At:                   at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating journal entry at")
	}

	return body, 0, nil
}

func CreateJournalEntryDetails(tx *gorm.DB, body *JournalEntryBody, at models.At) error {
	for i := range body.JournalEntryDetails {
		detail := &body.JournalEntryDetails[i]
		detail.JournalEntryId = body.JournalEntry.ID // assign FK to parent

		if err := services.DbInsert(tx, detail); err != nil {
			return errors.New("failed creating journal entry details")
		}

		// Audit trail for each detail
		atdataDetail := accounting_models.JournalEntryDetails2At{
			RefId:                       detail.ID,
			JournalEntryDetails2Content: detail.JournalEntryDetails2Content,
			At:                          at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating journal entry details at")
		}
	}
	return nil
}

func UpdateJournalEntry(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (interface{}, int, error) {
	var body JournalEntryBody

	//Parse full request
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	//Update main Journal Entry
	if err := services.DbUpdate(tx, &body.JournalEntry, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating journal entry")
	}

	//Get audit info
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	// Handle details
	if err := UpdateJournalEntryDetails(tx, &body, conditions, at); err != nil {
		return body.JournalEntry, fiber.StatusInternalServerError, err
	}

	//Audit record for main request
	atdata := accounting_models.JournalEntry2At{
		RefId:                body.JournalEntry.ID,
		JournalEntry2Content: body.JournalEntry.JournalEntry2Content,
		At:                   at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating journal entry at")
	}

	return body, 0, nil
}

func UpdateJournalEntryDetails(tx *gorm.DB, body *JournalEntryBody, conditions map[string]interface{}, at models.At) error {
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
		atdataDetail := accounting_models.JournalEntryDetails2At{
			RefId:                       detail.ID,
			JournalEntryDetails2Content: detail.JournalEntryDetails2Content,
			At:                          at,
		}

		if err := services.DbInsert(tx, &atdataDetail); err != nil {
			return errors.New("failed creating journal entry details at")
		}
	}
	return nil
}

func DeleteJournalEntry(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (interface{}, int, error) {
	var body JournalEntryBody

	//Parse full request
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	//Get audit info
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	//Delete main Journal Entry
	if err := services.DbDelete(tx, &body.JournalEntry, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting journal entry")
	}

	if err := DeleteJournalEntryDetails(tx, &body, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	//Audit record for main request
	atdata := accounting_models.JournalEntry2At{RefId: body.JournalEntry.ID, JournalEntry2Content: body.JournalEntry.JournalEntry2Content, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating journal entry at")
	}

	return body, 0, nil
}

func DeleteJournalEntryDetails(tx *gorm.DB, body *JournalEntryBody, at models.At) error {
	// Delete all details
	if err := services.DbDelete(tx, &accounting_models.JournalEntryDetails2{}, map[string]interface{}{"journal_entry_id": body.JournalEntry.ID}); err != nil {
		return errors.New("failed deleting all journal entry details")
	}

	// Optionally fetch deleted details for audit trail
	var deletedDetails []accounting_models.JournalEntryDetails2
	if err := tx.Unscoped().Where("journal_entry_id = ?", body.JournalEntry.ID).Find(&deletedDetails).Error; err == nil {
		for _, detail := range deletedDetails {
			atdataDetail := accounting_models.JournalEntryDetails2At{
				RefId:                       detail.ID,
				JournalEntryDetails2Content: detail.JournalEntryDetails2Content,
				At:                          at,
			}
			if err := services.DbInsert(tx, &atdataDetail); err != nil {
				return errors.New("failed creating journal entry details audit record")
			}
		}
	}
	return nil
}
