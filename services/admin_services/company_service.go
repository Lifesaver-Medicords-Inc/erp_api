package adminservices

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

type CompanyService struct {
}

func NewCompanyService() *CompanyService {
	return &CompanyService{}
}

func (c *CompanyService) CreateCompanyService(company *models.CompanyModel, at models.At) (*models.CompanyModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return company, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &company); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating company")
		}
		tx.Rollback()
		return company, fiber.StatusInternalServerError, err
	}

	atdata := models.CompanyAt{RefId: company.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return company, fiber.StatusInternalServerError, errors.New("failed creating companyat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return company, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return company, fiber.StatusCreated, nil
}

func (c *CompanyService) GetCompanyService(conditions map[string]interface{}) (*models.CompanyModel, int, error) {
	tx := initializers.DB.Begin()

	var company = &models.CompanyModel{}
	if tx.Error != nil {
		return company, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Preload("Address").Preload("Contacts").First(company).Error; err != nil {
		return company, fiber.StatusNotFound, errors.New("failed getting company")
	}

	return company, fiber.StatusOK, nil
}

func (c *CompanyService) GetCompaniesService(conditions map[string]interface{}) (*[]models.CompanyModel, int, error) {
	tx := initializers.DB.Begin()

	var companies = &[]models.CompanyModel{}

	if tx.Error != nil {
		return companies, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Preload("Address").Preload("Contacts").Find(companies).Error; err != nil {
		fmt.Println("ERROR:", err)
		return companies, fiber.StatusNotFound, errors.New("failed getting companies")
	}
	return companies, fiber.StatusOK, nil
}

func (c *CompanyService) UpdateCompanyService(company *models.CompanyModel, conditions map[string]interface{}, at models.At) (*models.CompanyModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return company, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &company, conditions); err != nil {
		return company, fiber.StatusInternalServerError, errors.New("failed updating company")
	}

	atdata := models.CompanyAt{RefId: company.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return company, fiber.StatusInternalServerError, errors.New("failed creating companyat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return company, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return company, fiber.StatusOK, nil
}

func (c *CompanyService) DeleteCompanyService(conditions map[string]interface{}, at models.At) (*models.CompanyModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.CompanyModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	company, status, err := c.GetCompanyService(conditions)

	if err != nil {
		return company, status, errors.New("company not found")
	}
	fmt.Println(company)
	fmt.Println(conditions)
	if err := services.DbDelete(tx, &company, conditions); err != nil {
		return company, fiber.StatusInternalServerError, errors.New("failed deleting company")
	}

	atdata := models.CompanyAt{RefId: company.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return company, fiber.StatusInternalServerError, errors.New("failed creating companyat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return company, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return company, fiber.StatusOK, nil
}

func (c *CompanyService) CreateCompanyAddressService(address *models.CompanyAddressModel, at models.At) (*models.CompanyAddressModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return address, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &address); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating address")
		}
		tx.Rollback()
		return address, fiber.StatusInternalServerError, err
	}

	atdata := models.CompanyAddressAt{RefId: address.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return address, fiber.StatusInternalServerError, errors.New("failed creating companyat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return address, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return address, fiber.StatusCreated, nil
}

func (c *CompanyService) CreateCompanyContactService(contact *models.CompanyContactModel, at models.At) (*models.CompanyContactModel, int, error) {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return contact, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &contact); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating contact")
		}
		tx.Rollback()
		return contact, fiber.StatusInternalServerError, err
	}

	atdata := models.CompanyContactAt{RefId: contact.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return contact, fiber.StatusInternalServerError, errors.New("failed creating contactat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return contact, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return contact, fiber.StatusCreated, nil
}

func (c *CompanyService) UpdateCompanyAddressService(address *models.CompanyAddressModel, conditions map[string]interface{}, at models.At) (*models.CompanyAddressModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return address, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &address, conditions); err != nil {
		return address, fiber.StatusInternalServerError, errors.New("failed updating company address")
	}

	atdata := models.CompanyAddressAt{RefId: address.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return address, fiber.StatusInternalServerError, errors.New("failed creating company address at")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return address, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return address, fiber.StatusOK, nil
}

func (c *CompanyService) UpdateCompanyContactService(contact *models.CompanyContactModel, conditions map[string]interface{}, at models.At) (*models.CompanyContactModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return contact, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbUpdate(tx, &contact, conditions); err != nil {
		return contact, fiber.StatusInternalServerError, errors.New("failed updating company contact")
	}

	atdata := models.CompanyContactAt{RefId: contact.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return contact, fiber.StatusInternalServerError, errors.New("failed creating company contact at")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return contact, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return contact, fiber.StatusOK, nil
}
