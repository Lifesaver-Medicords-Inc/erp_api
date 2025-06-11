package setup_services

import (
	// "errors"

	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetGeneralLedgerMappers(conditions map[string]interface{}) ([]models.GeneralLedgerMapper, int, error) {

	var based_service = services.NewInMemoryRepository(nil, nil, models.GeneralLedgerMapper{}, models.GeneralLedgerMapperAt{})

	return based_service.FetchAll()
}
func GetGeneralLedgerMapper(id int) (models.GeneralLedgerMapper, int, error) {

	conditions := map[string]interface{}{
		"id": id,
	}

	var Group models.GeneralLedgerMapper

	if err := services.DbGet(&Group, conditions); err != nil {
		return Group, fiber.StatusInternalServerError, errors.New("failed getting Group")
	}

	return Group, 0, nil
}

func UpdateGeneralLedgerMapper(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (int, error) {

	var based_service = services.NewInMemoryRepository(c, tx, models.GeneralLedgerMapper{}, models.GeneralLedgerMapperAt{})

	var payload models.GeneralLedgerMapperPayload

	if err := c.BodyParser(&payload); err != nil {
		fmt.Println("ERROR:")
		return fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	for index, value := range payload.Payload {
		fmt.Println("Index:", index, "Value:", value)

		at, ok := c.Locals("at").(models.At)
		if !ok {
			at = models.At{}
		}

		atdata := models.GeneralLedgerMapperAt{RefId: value.ID, PseudoAccount: value.PseudoAccount, AccountId: value.AccountId, At: at}

		based_service.Update(value, atdata, conditions)
	}

	return 200, nil
}
