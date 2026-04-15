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

func (fs *VehicleFileService) SaveUploadedFileService(file multipart.File, header *multipart.FileHeader, vehicleId int) (*models.VehicleFileModel, int, error) {
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
	return record, fiber.StatusOK, nil
}

func (fs *VehicleFileService) SaveVehicleFileService(file *models.VehicleFileModel, at models.At) (*models.VehicleFileModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.VehicleFileModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, &file); err != nil {
		// Remove file on failed
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

	return file, fiber.StatusOK, nil
}

func (v *VehicleFileService) GetVehicleFileService(conditions map[string]interface{}) (*models.VehicleFileModel, int, error) {
	tx := initializers.DB.Begin()

	var file = &models.VehicleFileModel{}

	if tx.Error != nil {
		return file, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).First(file).Error; err != nil {
		return file, fiber.StatusNotFound, errors.New("failed getting vehicle file")
	}

	return file, fiber.StatusOK, nil
}

func (v *VehicleFileService) GetVehicleFilesService(conditions map[string]interface{}) (*[]models.VehicleModel, int, error) {
	tx := initializers.DB.Begin()

	var vehicles = &[]models.VehicleModel{}

	if tx.Error != nil {
		return vehicles, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Where(conditions).Find(vehicles).Error; err != nil {
		return vehicles, fiber.StatusNotFound, errors.New("failed getting vehicles files")
	}
	return vehicles, fiber.StatusOK, nil
}

func (fs *VehicleFileService) RemoveVehicleFileService(conditions map[string]interface{}, at models.At) (*models.VehicleFileModel, int, error) {
	tx := initializers.DB.Begin()

	if tx.Error != nil {
		return &models.VehicleFileModel{}, fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	file, status, err := fs.GetVehicleFileService(conditions)

	if err != nil {
		return file, status, errors.New("vehicle not found")
	}

	if err := services.DbDelete(tx, &file, conditions); err != nil {
		return file, fiber.StatusInternalServerError, errors.New("failed deleting vehicle file")
	}

	os.Remove(file.FilePath)

	atdata := models.VehicleFileAt{RefId: file.ID, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		tx.Rollback()
		return file, fiber.StatusInternalServerError, errors.New("failed creating vehiclefileat")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return file, fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return file, fiber.StatusOK, nil
}
