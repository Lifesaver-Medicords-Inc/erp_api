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

func CreateItemImage(tx *gorm.DB, basedId uint, itemImage ItemImageInput, at models.At) error {

	for _, base64Image := range itemImage.Images {
		filePath, err := services.UploadFile(base64Image)
		if err != nil {
			return errors.New("failed to upload image locally")
		}

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

func UpdateItemImage(tx *gorm.DB, basedId uint, itemImage ItemImageInput, at models.At, conditions map[string]interface{}) error {
	var existingImages []models.ItemImage
	if err := tx.Where("based_id = ?", basedId).Find(&existingImages).Error; err != nil {
		return errors.New("failed to fetch existing item images")
	}

	existingImagePaths := make(map[string]uint)
	for _, img := range existingImages {
		existingImagePaths[img.Image] = img.ID
	}

	newImagePaths := make(map[string]bool)
	for _, base64Image := range itemImage.Images {
		filePath, err := services.UploadFile(base64Image)
		if err != nil {
			return errors.New("failed to upload image locally")
		}
		newImagePaths[filePath] = true

		if _, exists := existingImagePaths[filePath]; !exists {
			content := models.ItemImageContent{
				BasedId: basedId,
				Image:   filePath,
			}

			newItemImage := models.ItemImage{ItemImageContent: content}
			if err := services.DbInsert(tx, &newItemImage); err != nil {
				return errors.New("failed creating item image")
			}

			newItemImageAt := models.ItemImageAt{
				RefId:            newItemImage.ID,
				ItemImageContent: content,
				At:               at,
			}
			if err := services.DbInsert(tx, &newItemImageAt); err != nil {
				return errors.New("failed creating item image at")
			}
		}
	}

	for path, imgID := range existingImagePaths {
		if _, exists := newImagePaths[path]; !exists {
			if err := tx.Where("id = ?", imgID).Delete(&models.ItemImage{}).Error; err != nil {
				return errors.New("failed to delete old item image")
			}
			if err := tx.Where("ref_id = ?", imgID).Delete(&models.ItemImageAt{}).Error; err != nil {
				return errors.New("failed to delete old item image at")
			}
		}
	}

	return nil
}
