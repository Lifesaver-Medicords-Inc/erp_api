package adminhandlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
	"github.com/pierceperado/smpc/utils"
)

type CompanyHandler struct {
	CompanyService *adminservices.CompanyService
}

func NewCompanyHandler(service *adminservices.CompanyService) *CompanyHandler {
	return &CompanyHandler{
		CompanyService: service,
	}
}

func (co *CompanyHandler) CreateCompanyHandler(c *fiber.Ctx) error {
	var body models.CompanyModel

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	// Detached before the company row is inserted - see UpdateCompanyHandler. GORM inserted
	// them with the company, then the code below inserted them again: every new company got
	// its address and each contact twice.
	address := body.Address
	contacts := body.Contacts
	body.Address = models.CompanyAddressModel{}
	body.Contacts = nil

	data, status, err := co.CompanyService.CreateCompanyService(&body, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	if address.CompanyAddressContent != (models.CompanyAddressContent{}) {
		address.ID = 0
		address.CompanyId = data.ID
		if _, _, err := co.CompanyService.CreateCompanyAddressService(&address, at); err != nil {
			return utils.RespondError(c, fiber.StatusInternalServerError, "Company created, but its address was not: "+err.Error())
		}
	}

	for _, contact := range contacts {
		contact.ID = 0
		contact.CompanyId = data.ID
		if _, _, err := co.CompanyService.CreateCompanyContactService(&contact, at); err != nil {
			return utils.RespondError(c, fiber.StatusInternalServerError, "Company created, but a contact was not: "+err.Error())
		}
	}

	if saved, _, err := co.CompanyService.GetCompanyService(map[string]interface{}{"id": data.ID}); err == nil {
		data = saved
	}

	return utils.RespondSuccess(c, data)
}

func (co *CompanyHandler) GetCompanyHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	position, status, err := co.CompanyService.GetCompanyService(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, position)
}

func (co *CompanyHandler) GetCompaniesHandler(c *fiber.Ctx) error {
	id := c.Query("id")
	name := c.Query("company-name")
	status := c.Query("status")

	conditions := make(map[string]interface{})

	idNum, _ := strconv.Atoi(id)

	if idNum != 0 {
		conditions["id"] = id
	}

	if name != "" {
		conditions["company_name"] = name
	}

	if status != "" {
		conditions["status"] = status
	}

	companies, code, err := co.CompanyService.GetCompaniesService(conditions)
	if err != nil {
		return utils.RespondError(c, code, err.Error())
	}
	return utils.RespondSuccess(c, companies)
}

func (co *CompanyHandler) UpdateCompanyHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")

	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	var body models.CompanyModel

	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}
	// The record is the one in the URL; the address and contacts below hang off this id.
	body.ID = uint(idNum)

	// Detached before the company row is written. GORM saves a model's associations along
	// with it, so an address or contact without an id was INSERTED by the company update
	// itself - a second address row on every save - and then written again below.
	address := body.Address
	contacts := body.Contacts
	body.Address = models.CompanyAddressModel{}
	body.Contacts = nil

	data, status, err := co.CompanyService.UpdateCompanyService(&body, conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	// Address and contacts. An update used to write only to rows that already existed: the
	// address by company_id, a contact by its id. A company saved without an address (company 1
	// has none) could never be given one, and a contact added while editing (id 0) matched no
	// row and was silently dropped. Now: the address is updated when it exists and created when
	// it does not and something was entered; a contact with an id is updated within this
	// company, one without is created. A caller that sends neither (Accounting's Company Setup)
	// changes neither - an empty address is never created.
	address.CompanyId = data.ID
	hasAddress, err := co.CompanyService.CompanyHasAddress(data.ID)
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Company saved, but its address could not be checked: "+err.Error())
	}
	if hasAddress {
		if _, _, err := co.CompanyService.UpdateCompanyAddressService(&address, map[string]interface{}{"company_id": data.ID}, at); err != nil {
			return utils.RespondError(c, fiber.StatusInternalServerError, "Company saved, but its address was not: "+err.Error())
		}
	} else if address.CompanyAddressContent != (models.CompanyAddressContent{}) {
		address.ID = 0
		if _, _, err := co.CompanyService.CreateCompanyAddressService(&address, at); err != nil {
			return utils.RespondError(c, fiber.StatusInternalServerError, "Company saved, but its address was not: "+err.Error())
		}
	}

	for _, contact := range contacts {
		contact.CompanyId = data.ID
		if contact.ID == 0 {
			if _, _, err := co.CompanyService.CreateCompanyContactService(&contact, at); err != nil {
				return utils.RespondError(c, fiber.StatusInternalServerError, "Company saved, but a new contact was not: "+err.Error())
			}
			continue
		}
		condition := map[string]interface{}{"id": contact.ID, "company_id": data.ID}
		if _, _, err := co.CompanyService.UpdateCompanyContactService(&contact, condition, at); err != nil {
			return utils.RespondError(c, fiber.StatusInternalServerError, "Company saved, but a contact was not: "+err.Error())
		}
	}

	// The company as stored, address and contacts included, so the caller shows what was saved.
	if saved, _, err := co.CompanyService.GetCompanyService(map[string]interface{}{"id": data.ID}); err == nil {
		data = saved
	}

	return utils.RespondSuccess(c, data)
}

func (co *CompanyHandler) DeleteCompanyHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)

	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}
	fmt.Println(conditions)
	data, status, err := co.CompanyService.DeleteCompanyService(conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
