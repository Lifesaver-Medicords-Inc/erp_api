package setup_services

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetItemImages(itemImage *[]models.ItemImage, conditions map[string]interface{}) error {
	if err := services.DbGet(itemImage, conditions); err != nil {
		fmt.Println("ERROR:", err)
		return errors.New("failed getting item images")
	}

	for _, img := range *itemImage {
		for _, field := range []string{"Img1", "Img2", "Img3", "Img4", "Img5", "Img6"} {
			fieldValue := reflect.ValueOf(&img).Elem().FieldByName(field)

			if fieldValue.IsValid() && fieldValue.Kind() == reflect.String {
				imgPath := fieldValue.String()
				if imgPath != "" {
					base64Str, err := EncodeFileToBase64(imgPath)
					if err != nil {
						fmt.Println("ERROR:", err)
						return errors.New("failed converting image to base64")
					}
					fieldValue.SetString(base64Str)
				}
			} else {
				// If no image path, leave as empty string (for missing images)
				fieldValue.SetString("")
			}
		}
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

func CreateItemImage(tx *gorm.DB, basedId uint, itemImage models.ItemImage, at models.At) error {
	imgPaths := make(map[string]string)

	imgFields := []string{"img1", "img2", "img3", "img4", "img5", "img6"}
	imgValues := []string{itemImage.Img1, itemImage.Img2, itemImage.Img3, itemImage.Img4, itemImage.Img5, itemImage.Img6}

	for i, field := range imgFields {
		if imgValues[i] != "" { // Only process if an image is provided
			var err error
			imgPaths[field], err = services.UploadFile(imgValues[i])
			if err != nil {
				return fmt.Errorf("error saving %s: %v", field, err)
			}
		}
	}

	content := models.ItemImageContent{
		BasedId: basedId,
		Img1:    imgPaths["img1"],
		Img2:    imgPaths["img2"],
		Img3:    imgPaths["img3"],
		Img4:    imgPaths["img4"],
		Img5:    imgPaths["img5"],
		Img6:    imgPaths["img6"],
	}

	itemImg := models.ItemImage{ItemImageContent: content}
	if err := services.DbInsert(tx, &itemImg); err != nil {
		return errors.New("failed creating item image")
	}

	itemImgAt := models.ItemImageAt{
		RefId:            itemImg.ID,
		ItemImageContent: content,
		At:               at,
	}

	if err := services.DbInsert(tx, &itemImgAt); err != nil {
		return errors.New("failed creating item image at")
	}

	return nil
}
func UpdateItemImage(tx *gorm.DB, itemImage models.ItemImage, at models.At, conditions map[string]interface{}) error {
	imgPaths := make(map[string]string)

	imgFields := []string{"img1", "img2", "img3", "img4", "img5", "img6"}
	imgValues := []string{itemImage.Img1, itemImage.Img2, itemImage.Img3, itemImage.Img4, itemImage.Img5, itemImage.Img6}

	for i, field := range imgFields {
		if imgValues[i] != "" { // Only process if an image is provided
			var err error
			imgPaths[field], err = services.UploadFile(imgValues[i])
			if err != nil {
				return fmt.Errorf("error saving %s: %v", field, err)
			}
		}
	}

	updatedContent := models.ItemImageContent{
		BasedId: itemImage.BasedId,
		Img1:    imgPaths["img1"],
		Img2:    imgPaths["img2"],
		Img3:    imgPaths["img3"],
		Img4:    imgPaths["img4"],
		Img5:    imgPaths["img5"],
		Img6:    imgPaths["img6"],
	}

	if err := services.DbUpdate(tx, &updatedContent, conditions); err != nil {
		fmt.Println("ERROR:", err)
		return errors.New("failed updating item image")
	}

	itemImgAt := models.ItemImageAt{
		RefId:            itemImage.ID,
		ItemImageContent: updatedContent,
		At:               at,
	}
	if err := services.DbInsert(tx, &itemImgAt); err != nil {
		return errors.New("failed creating item image at")
	}

	return nil
}
