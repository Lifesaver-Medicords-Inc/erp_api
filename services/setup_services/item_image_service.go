package setup_services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

func CreateItemImage(tx *gorm.DB, basedId uint, imageInput ItemImageInput, at models.At) error {

	for _, base64Image := range imageInput.Images {
		filePath, err := services.UploadFile(base64Image)
		if err != nil {
			return errors.New("failed to upload image locally")
		}

		// Prepare content for database
		content := models.ItemImageContent{
			BasedId: basedId,
			Image:   filePath,
		}

		itemImage := models.ItemImage{ItemImageContent: content}
		if err := services.DbInsert(tx, &itemImage); err != nil {
			return errors.New("failed creating item image")
		}

		itemImageAt := models.ItemImageAt{
			RefId:            itemImage.ID,
			ItemImageContent: content,
			At:               at,
		}
		if err := services.DbInsert(tx, &itemImageAt); err != nil {
			return errors.New("failed creating item image at")
		}
	}

	return nil
}
