package setup_services

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetItemImages(itemImage *[]models.ItemImage, conditions map[string]interface{}) error {
	if err := services.DbGet(itemImage, conditions); err != nil {
		return errors.New("failed getting itemimage")
	}

	return nil
}

func EncodeFileToBase64(filePath string) (string, error) {
	// Read the file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed reading file: %w", err)
	}

	// Encode the file data to base64
	base64Str := base64.StdEncoding.EncodeToString(fileData)
	return base64Str, nil
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
 