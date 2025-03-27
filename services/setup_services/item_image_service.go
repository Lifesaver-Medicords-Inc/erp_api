package setup_services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetImageURL(filename string) string {
	host := "localhost"
	port := os.Getenv("BIND_PORT")

	return fmt.Sprintf("%s:%s/files/%s", host, port, filename)
}

func GetItemImages(itemImage *[]models.ItemImage, conditions map[string]interface{}) error {
	if err := services.DbGet(itemImage, conditions); err != nil {
		return errors.New("failed getting item images")
	}

	for i := range *itemImage {
		(*itemImage)[i].Image = GetImageURL(filepath.Base((*itemImage)[i].Image))
	}

	return nil
}

type ImageBody struct {
	BasedId    uint     `json:"based_id"`
	ItemImages []string `json:"itemimages"`
}

func CreateItemImage(c *fiber.Ctx, tx *gorm.DB) (ImageBody, int, error) {
	var body ImageBody
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	for _, base64Image := range body.ItemImages {
		filePath, err := services.UploadFile(base64Image)
		if err != nil {
			return body, fiber.StatusInternalServerError, errors.New("failed to upload image locally")
		}

		content := models.ItemImageContent{
			BasedId: body.BasedId,
			Image:   filePath,
		}

		itemimage := models.ItemImage{ItemImageContent: content}
		if err := services.DbInsert(tx, &itemimage); err != nil {
			return body, fiber.StatusInternalServerError, errors.New("failed creating item images")
		}

		at, ok := c.Locals("at").(models.At)
		if !ok {
			at = models.At{}
		}

		itemimageat := models.ItemImageAt{
			RefId:            itemimage.ID,
			ItemImageContent: itemimage.ItemImageContent,
			At:               at,
		}

		if err := services.DbInsert(tx, &itemimageat); err != nil {
			return body, fiber.StatusInternalServerError, errors.New("failed creating itemimageat")
		}
	}

	return body, 0, nil
}

func UpdateItemImage(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.ItemImage, int, error) {
	var body models.ItemImage

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	// // Fetch the existing image record before deletion
	// var existingImage models.ItemImage
	// if err := tx.Where(conditions).First(&existingImage).Error; err != nil {
	// 	return body, fiber.StatusNotFound, errors.New("image not found")
	// }

	// // Delete the image file from storage
	// if err := services.DeleteFile(existingImage.Image); err != nil {
	// 	return body, fiber.StatusInternalServerError, errors.New("failed to delete image file")
	// }

	filePath, err := services.UploadFile(body.Image)
	if err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed to upload image locally")
	}

	body.Image = filePath


	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		fmt.Println("update body:", body)
		return body, fiber.StatusInternalServerError, errors.New("failed updating itemimage")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	itemimageat := models.ItemImageAt{
		RefId:            body.ID,
		ItemImageContent: body.ItemImageContent,
		At:               at,
	}

	if err := services.DbInsert(tx, &itemimageat); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating itemimageat")
	}

	return body, 0, nil
}

func DeleteItemImage(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.ItemImage, int, error) {
	var body models.ItemImage

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	// // Fetch the existing image record before deletion
	// var existingImage models.ItemImage
	// if err := tx.Where(conditions).First(&existingImage).Error; err != nil {
	// 	return body, fiber.StatusNotFound, errors.New("image not found")
	// }

	// // Delete the image file from storage
	// if err := services.DeleteFile(existingImage.Image); err != nil {
	// 	return body, fiber.StatusInternalServerError, errors.New("failed to delete image file")
	// }

	// Delete the record from the database
	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting item image")
	}

	// Log the deletion (audit tracking)
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	itemimageat := models.ItemImageAt{
		RefId:            body.ID,
		ItemImageContent: body.ItemImageContent,
		At:               at,
	}

	if err := services.DbInsert(tx, &itemimageat); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating itemimageat")
	}

	return body, fiber.StatusOK, nil
}
