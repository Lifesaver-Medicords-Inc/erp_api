package bpi_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// ─── CreateBpiEntity ──────────────────────────────────────────────────────────

func CreateBpiEntity(tx *gorm.DB, parentId uint, entityId uint, salesId string, at models.At) error {
	fmt.Println("CREATE BPI ENTITY")

	content := models.BpiEntityContent{
		BpiGeneralId: parentId,
		EntityId:     entityId,
	}
	bpiEntity := models.BpiEntity{BpiEntityContent: content}

	if err := services.DbInsert(tx, &bpiEntity); err != nil {
		return errors.New("failed creating bpi entity")
	}

	bpiEntityAt := models.BpiEntityAt{
		RefId:            bpiEntity.BpiGeneralId,
		BpiEntityContent: content,
		At:               at,
	}
	if err := services.DbInsert(tx, &bpiEntityAt); err != nil {
		return errors.New("failed creating bpi entity_at")
	}

	// Fetch entity master to check its code
	var entity models.Entity // adjust to your actual Entity model
	if err := tx.First(&entity, entityId).Error; err != nil {
		return errors.New("failed to find entity")
	}

	// Fetch current general to check if number already exists
	var gen models.BpiGeneral
	if err := tx.First(&gen, parentId).Error; err != nil {
		return errors.New("failed to find bpi general")
	}

	switch strings.ToUpper(entity.Code) {
	case "CUS":
		// Only generate if not yet assigned — reuse existing number if re-selected
		if gen.CustomerCode == "" {
			if err := generateCustomerCode(tx, parentId); err != nil {
				return err
			}
		}
	case "SUP":
		// Only generate if not yet assigned — reuse existing number if re-selected
		if gen.SupplierCode == "" {
			if err := generateSupplierCode(tx, parentId); err != nil {
				return err
			}
		}
	}

	return nil
}

// ─── UpdateBpiEntity ──────────────────────────────────────────────────────────

func UpdateBpiEntity(tx *gorm.DB, parentId uint, entityId uint, salesId string, at models.At) error {
	conditions := map[string]interface{}{
		"bpi_general_id": parentId,
		"entity_id":      entityId,
	}

	fmt.Println("ENTITY ID", entityId)

	content := models.BpiEntityContent{
		EntityId: entityId,
	}

	bpiEntity := models.BpiEntity{BpiEntityContent: content}
	if err := services.DbUpdate(tx, &bpiEntity, conditions); err != nil {
		return errors.New("failed updating bpi entity")
	}

	bpiEntityAt := models.BpiEntityAt{
		RefId:            bpiEntity.BpiGeneralId,
		BpiEntityContent: content,
		At:               at,
	}
	if err := services.DbInsert(tx, &bpiEntityAt); err != nil {
		return errors.New("failed creating bpi_industries_at")
	}

	if err := CreateBpiHistory(tx, parentId, "update", "General Entities", salesId, at); err != nil {
		return err
	}

	return nil
}

// ─── Number Generators ────────────────────────────────────────────────────────

// Bug #292 (Trello): codes weren't incrementing - COUNT(*) of existing
// non-blank codes stands in for "the next number" here, but a delete or an
// entity type toggled off-then-back-on drops the count below the highest
// number actually issued, so the next COUNT+1 collides with (or trails) an
// already-issued code. Read the highest number ever issued instead, so a
// gap in the sequence never gets reused.
func generateCustomerCode(tx *gorm.DB, bpiGeneralId uint) error {
	var maxNum int
	if err := tx.Raw(`
		SELECT ISNULL(MAX(TRY_CAST(SUBSTRING(customer_code, 3, LEN(customer_code)) AS INT)), 0)
		FROM tbl_bpi_general
		WHERE customer_code LIKE 'C#%'
	`).Scan(&maxNum).Error; err != nil {
		return errors.New("failed computing next customer_code")
	}

	customerCode := fmt.Sprintf("C#%04d", maxNum+1)

	if err := tx.Model(&models.BpiGeneral{}).
		Where("id = ?", bpiGeneralId).
		Update("customer_code", customerCode).Error; err != nil {
		return errors.New("failed setting customer_code")
	}

	fmt.Println("Generated CustomerCode:", customerCode)
	return nil
}

// Bug #292 (Trello): same fix as generateCustomerCode above.
func generateSupplierCode(tx *gorm.DB, bpiGeneralId uint) error {
	var maxNum int
	if err := tx.Raw(`
		SELECT ISNULL(MAX(TRY_CAST(SUBSTRING(supplier_code, 3, LEN(supplier_code)) AS INT)), 0)
		FROM tbl_bpi_general
		WHERE supplier_code LIKE 'S#%'
	`).Scan(&maxNum).Error; err != nil {
		return errors.New("failed computing next supplier_code")
	}

	supplierCode := fmt.Sprintf("S#%04d", maxNum+1)

	if err := tx.Model(&models.BpiGeneral{}).
		Where("id = ?", bpiGeneralId).
		Update("supplier_code", supplierCode).Error; err != nil {
		return errors.New("failed setting supplier_code")
	}

	fmt.Println("Generated SupplierCode:", supplierCode)
	return nil
}
