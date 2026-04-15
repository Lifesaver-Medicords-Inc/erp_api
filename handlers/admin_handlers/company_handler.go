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

	data, status, err := co.CompanyService.CreateCompanyService(&body, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	address := body.Address
	contacts := body.Contacts

	address.CompanyId = data.ID

	a, _, err := co.CompanyService.CreateCompanyAddressService(&address, at)

	if err != nil {
		data.Address = *a
	}

	for _, contact := range contacts {
		contact.CompanyId = data.ID
		con, _, err := co.CompanyService.CreateCompanyContactService(&contact, at)

		if err != nil {
			contacts = append(contacts, *con)
		}
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

	data, status, err := co.CompanyService.UpdateCompanyService(&body, conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	address := body.Address
	contacts := body.Contacts

	updateAddressCondition := map[string]interface{}{
		"company_id": data.ID,
	}

	a, _, err := co.CompanyService.UpdateCompanyAddressService(&address, updateAddressCondition, at)

	if err != nil {
		data.Address = *a
	}

	for _, contact := range contacts {
		condition := map[string]interface{}{
			"id": contact.ID,
		}
		fmt.Println(contact)

		con, _, err := co.CompanyService.UpdateCompanyContactService(&contact, condition, at)

		if err != nil {
			contacts = append(contacts, *con)
		}
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
