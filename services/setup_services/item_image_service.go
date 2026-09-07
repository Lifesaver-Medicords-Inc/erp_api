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

// removeStoredImage deletes an item image's file from the files directory.
//
// A file that is already gone is NOT an error. The database row is the record of the
// image; the file is a copy of it on disk, and the two can legitimately be out of step -
// a restored database is the obvious case (the rows come across, the files folder does
// not), and so is a file removed by hand or a half-finished earlier delete.
//
// Treating a missing file as fatal made those images undeletable forever: the error
// aborted the enclosing transaction, so the whole item update rolled back and the row
// stayed exactly where it was, to fail again on the next attempt (user-reported
// 2026-09-05, "failed to delete image file" on Item Entry). Every image row in the
// rehearsal database points at a file that is not on disk, so this was every image.
//
// Same rule job_order_service.go already applies when replacing a report file.
//
// Path comes from FilesDir() rather than a hard-coded "files/" so it still resolves if
// FILES_DIR is set - which is exactly what a real server deployment would do.
func removeStoredImage(image string) error {
	if image == "" {
		return nil
	}

	// Stored as a bare file name (UploadFile returns one, and GetItemImages re-bases what
	// it hands back), but Base() also makes this safe against a legacy row that kept a
	// full path.
	path := filepath.Join(services.FilesDir(), filepath.Base(image))

	if err := services.DeleteFile(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func GetImageURL(filename string) string {
	return filename
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

func CreateItemImageChild(tx *gorm.DB, basedId uint, newImages []models.ItemImage, at models.At) error {
	for _, images := range newImages {
		filePath, err := services.UploadFile(images.Image)
		if err != nil {
			fmt.Println("CREATE ERR:", err)
			return errors.New("failed to upload image locally")
		}

		content := models.ItemImageContent{
			BasedId:  basedId,
			Image:    filePath,
			Filename: images.Filename,
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

func UpdateItemImageChild(tx *gorm.DB, itemImageID uint, newImage string, filename string, at models.At) error {
	// Find existing image
	var body models.ItemImage
	if err := tx.Where("id = ?", itemImageID).First(&body).Error; err != nil {
		return errors.New("image not found")
	}

	// Delete old file. A missing one does not stop the replacement - see removeStoredImage.
	if err := removeStoredImage(body.Image); err != nil {
		fmt.Println("REPLACE ERR:", err)
		return errors.New("failed to delete old image file")
	}

	// Upload new image
	filePath, err := services.UploadFile(newImage)
	if err != nil {
		return errors.New("failed to upload new image")
	}

	// Update database record
	updateData := map[string]interface{}{
		"image":    filePath,
		"filename": filename,
	}
	if err := tx.Model(&models.ItemImage{}).Where("id = ?", itemImageID).Updates(updateData).Error; err != nil {
		return errors.New("failed updating item image")
	}

	// Log change in ItemImageAt
	itemImageAt := models.ItemImageAt{
		RefId: itemImageID,
		ItemImageContent: models.ItemImageContent{
			Image:    filePath,
			Filename: filename,
		},
		At: at,
	}

	if err := services.DbInsert(tx, &itemImageAt); err != nil {
		return errors.New("failed creating item image at")
	}

	return nil
}

func DeleteItemImageChild(tx *gorm.DB, itemImageID uint) error {
	var body models.ItemImage
	if err := tx.Where("id = ?", itemImageID).First(&body).Error; err != nil {
		fmt.Println("DELETE ERR:", err)
		return errors.New("image not found")
	}

	// The row goes regardless of whether its file is still on disk - see removeStoredImage.
	if err := removeStoredImage(body.Image); err != nil {
		fmt.Println("BODY IMAGE", body.Image, err)
		return errors.New("failed to delete image file")
	}

	if err := tx.Where("id = ?", itemImageID).Delete(&models.ItemImage{}).Error; err != nil {
		fmt.Println("DELETION ERROR:", err)
		return errors.New("failed deleting item image")
	}

	if err := tx.Where("ref_id = ?", itemImageID).Delete(&models.ItemImageAt{}).Error; err != nil {
		return errors.New("failed deleting item image at")
	}

	return nil
}

func UpdateItemImage(tx *gorm.DB, basedId uint, itemImage ItemImage, at models.At, conditions map[string]interface{}) error {
	// Create new images
	if len(itemImage.NewImages) > 0 {
		if err := CreateItemImageChild(tx, basedId, itemImage.NewImages, at); err != nil {
			return err
		}
	}

	// Replace images
	for _, replaceReq := range itemImage.ReplaceImages {
		fmt.Println("REPLACE REQ body", replaceReq)
		if err := UpdateItemImageChild(tx, replaceReq.ID, replaceReq.Image, replaceReq.Filename, at); err != nil {
			fmt.Println("REPLACE REQ", err)
			return err
		}
	}

	// Delete images
	for _, deleteReq := range itemImage.DeleteImages {
		if err := DeleteItemImageChild(tx, deleteReq.ID); err != nil {
			fmt.Println("DELETE REQ", deleteReq)
			return err
		}
	}

	return nil
}
