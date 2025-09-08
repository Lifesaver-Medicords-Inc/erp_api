package adminservices

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

type VehicleFileService struct {
	UploadDir string
}

func NewVehicleFileService(uploadDir string) *VehicleFileService {
	return &VehicleFileService{
		UploadDir: uploadDir,
	}
}

func (fs *VehicleFileService) SaveUploadedFile(file multipart.File, header *multipart.FileHeader, vehicleId int) (*models.VehicleFileModel, int, error) {

	ext := filepath.Ext(header.Filename)
	newFileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	savePath := filepath.Join(fs.UploadDir, newFileName)

	err := os.MkdirAll(fs.UploadDir, os.ModePerm)
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed to save file")
	}

	dst, err := os.Create(savePath)
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed to save file")
	}
	defer dst.Close()

	// Copy file
	_, err = io.Copy(dst, file)
	if err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed to save file")
	}

	// Detect file MIME type
	buf := make([]byte, 512)
	dst.Seek(0, io.SeekStart)
	n, _ := dst.Read(buf)
	fileType := http.DetectContentType(buf[:n])

	record := &models.VehicleFileModel{
		VehicleId: uint(vehicleId),
		VehicleFileContent: models.VehicleFileContent{
			FileName:     newFileName,
			OriginalName: header.Filename,
			FilePath:     savePath,
			Type:         fileType,
			Size:         int(header.Size),
		},
	}
	return record, 0, nil
}

func (fs *VehicleFileService) SaveVehicleFile(file *models.VehicleFileModel, at models.At) (*models.VehicleFileModel, int, error) {

	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.VehicleFileModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &file); err != nil {

		//Remove file on failed
		if file.FilePath != "" {
			os.Remove(file.FilePath)
		}

		if strings.Contains(err.Error(), "duplicate key") {

			err = errors.New("duplicate record error")
		} else {

			err = errors.New("failed saving file")
		}
		tx.Rollback()
		return file, fiber.StatusInternalServerError, err
	}

	atdata := models.VehicleAt{RefId: file.ID, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return file, fiber.StatusInternalServerError, errors.New("saving fileat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return file, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return file, 0, nil
}
