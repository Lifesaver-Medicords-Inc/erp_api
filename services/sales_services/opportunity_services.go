package sales_services

import (
	"errors"
	"fmt"

	//fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// test
func GetOpportunities(conditions map[string]interface{}) ([]models.OpportunityView, int, error) {

	var opportunities []models.OpportunityView

	if err := services.DbGet(&opportunities, conditions); err != nil {
		fmt.Println("ERROR:", err.Error())
		return opportunities, fiber.StatusInternalServerError, errors.New("failed getting opportunities")
	}
	fmt.Println("DATA: ", opportunities)
	return opportunities, 0, nil
}

func GetOpportunity(id int) (models.OpportunityView, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var opportunity models.OpportunityView

	if err := services.DbGet(&opportunity, conditions); err != nil {
		return opportunity, fiber.StatusInternalServerError, errors.New("failed getting opportunity")
	}

	return opportunity, 0, nil
}

// test
func CreateOpportunity(c *fiber.Ctx, tx *gorm.DB) (models.Opportunity, int, error) {
	var body models.Opportunity
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating opportunity")
		}
		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.OpportunityAt{RefId: body.ID, OpportunityContent: body.OpportunityContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating opportunityat")
	}
	key := services.GetKey(models.OpportunityView{}, nil)
	services.InvalidateCache(key)
	return body, 0, nil
}

func UpdateOpportunity(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.Opportunity, int, error) {
	var body models.Opportunity
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	fmt.Println("Body:", body)

	conditions = map[string]interface{}{
		"document_no": body.DocumentNo,
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating opportunity")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.OpportunityAt{RefId: body.ID, OpportunityContent: body.OpportunityContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating opportunityat")
	}
	fmt.Println("body: ", body)
	key := services.GetKey(models.OpportunityView{}, nil)
	services.InvalidateCache(key)
	return body, 0, nil
}
