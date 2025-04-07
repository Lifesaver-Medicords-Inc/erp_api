package bpi_services

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func saveBase64Image(base64Str string) (string, error) {
	if base64Str == "" {
		return "", nil
	}
	file, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return "", errors.New("failed decoding data")
	}
	if err := os.MkdirAll("./files", os.ModePerm); err != nil {
		return "", errors.New("failed creating folder")
	}

	fileName := time.Now().UnixNano()
	mimeType := mimetype.Detect(file)
	fileExtension := mimeType.Extension()

	path := fmt.Sprintf("./files/%d%v", fileName, fileExtension)
	if err := os.WriteFile(path, file, 0644); err != nil {
		return "", errors.New("failed saving file")
	}

	return path, nil
}

func CreateBpiAccreditation(tx *gorm.DB, parentId uint, general_id uint, child models.BpiAccreditation, at models.At) error {

	child.BpiAccreditationContent.BasedId = parentId
	child.BpiAccreditationContent.BranchId = general_id

	if !strings.HasPrefix(child.FilePath, "./") {
		path, err := saveBase64Image(child.BpiAccreditationContent.FilePath)

		if err != nil {
			return errors.New("failed to to convert file to base64")
		}
		child.BpiAccreditationContent.FilePath = path
		if err := services.DbInsert(tx, &child); err != nil {
			return errors.New("failed to create bpi accreditation")
		}

		childAt := models.BpiAccreditationAt{
			RefId:                   child.ID,
			BpiAccreditationContent: child.BpiAccreditationContent,
			At:                      at,
		}
		if err := services.DbInsert(tx, &childAt); err != nil {
			return errors.New("failed creating accreditation at")
		}
	}

	return nil
}

func UpdateBpiAccreditation(tx *gorm.DB, child models.BpiAccreditation, at models.At, parentId uint) error {

	conditions := map[string]interface{}{
		"based_id": parentId,
	}
	path, err := saveBase64Image(child.BpiAccreditationContent.FilePath)
	if err != nil {
		return errors.New("failed to to convert file to base64")
	}
	child.BpiAccreditationContent.FilePath = path
	if err := services.DbUpdate(tx, &child, conditions); err != nil {
		return errors.New("failed updating bpi accreditations")
	}
	childAt := models.BpiAccreditationAt{
		RefId:                   child.ID,
		BpiAccreditationContent: child.BpiAccreditationContent,
		At:                      at,
	}
	if err := services.DbInsert(tx, &childAt); err != nil {
		return errors.New("failed creating bpi accreditation at")
	}

	return nil
}
