package sales_services

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// retrieves a list of models in Application objects from the database
func GetApplications() ([]models.Application, error) {

	// using a slice to hold the results
	var applications []models.Application

	if err := services.DbGet(&applications, nil); err != nil {
		return applications, err
	}

	return applications, nil
}

// retrives a single data from the database based on the id
func GetApplication(id int) (models.Application, error) {
	condiions := map[string]interface{}{
		"id": id,
	}
	var application models.Application

	if err := services.DbGet(&application, condiions); err != nil {
		return application, err
	}

	return application, nil
}

// Create data
func CreateApplication(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.Application
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	if err := services.DbInsert(tx, &body); err != nil {
		return err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.ApplicationAt{
		RefId: body.ID,
		Code:  body.Code,
		Name:  body.Name,
		At:    at,
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}

// updates the data on the db
func UpdateApplication(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.Application
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	if err := services.DbUpdate(tx, &body, nil); err != nil {
		return err
	}

	atdata := models.ApplicationAt{RefId: body.ID, Code: body.Code, At: models.At{}}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}
	return nil
}

//
// Deletes the data on the database
//

func DeleteApplication(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.Application

	if err := c.BodyParser(&body); err != nil {
		return err
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return err
	}

	//at, ok := c.Locals("at").(models.At)
	//if !ok {
	//	return errors.New("error AT data")
	//}

	atdata := models.ApplicationAt{
		RefId: body.ID,
		Code:  body.Code,
		Name:  body.Name,
		At:    models.At{},
	}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}
