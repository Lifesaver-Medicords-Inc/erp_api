package setup_services

import (
	// "errors"

	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetBooks(conditions map[string]interface{}) ([]accounting_models.Book, int, error) {

	var based_service = services.NewInMemoryRepository(nil, nil, accounting_models.Book{}, accounting_models.BookAt{})

	return based_service.FetchAll()
}

func GetBook(conditions map[string]interface{}) ([]accounting_models.Book, int, error) {

	var based_service = services.NewInMemoryRepository(nil, nil, accounting_models.Book{}, accounting_models.BookAt{})

	return based_service.FetchWithFilter(conditions)
}

func CreateBook(c *fiber.Ctx, tx *gorm.DB) (accounting_models.Book, int, error) {

	var service = services.NewInMemoryRepository(c, tx, accounting_models.Book{}, accounting_models.BookAt{})

	var body accounting_models.Book
	if err := c.BodyParser(&body); err != nil {

		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := accounting_models.BookAt{RefId: body.ID, Code: body.Code, At: at}

	return service.Create(body, atdata)
}

func UpdateBook(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (accounting_models.Book, int, error) {

	var service = services.NewInMemoryRepository(c, tx, accounting_models.Book{}, accounting_models.BookAt{})

	var body accounting_models.Book
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := accounting_models.BookAt{RefId: body.ID, Code: body.Code, At: at}

	return service.Update(body, atdata, conditions)
}

func DeleteBook(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (accounting_models.Book, int, error) {

	var body accounting_models.Book
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting class")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := accounting_models.BookAt{RefId: body.ID, Code: body.Code, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating classat")
	}

	return body, 0, nil
}
