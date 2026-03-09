package setup_handlers

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/handlers/purchasing_handlers"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils" // Correct import path for excelize
	"github.com/xuri/excelize/v2"
)

func GetItems(c *fiber.Ctx) error {
	data, status, err := setup_services.GetItems(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetItem(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	data, status, err := setup_services.GetItem(idNum)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func CreateItem(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}
	data, status, err := setup_services.CreateItem(c, tx)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	//go broadcastItems() -- if used invalidate cache?
	//purchasing_handlers.BroadcastRedboxList()

	return utils.RespondSuccess(c, data)
}

func UpdateItem(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := setup_services.UpdateItem(c, tx, nil)

	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	purchasing_handlers.BroadcastRedboxList()

	return utils.RespondSuccess(c, data)
}

func DeleteItem(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := setup_services.DeleteItem(c, tx, nil)

	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func WsgetItems(c *websocket.Conn) {
	initializers.WM.AddClient(c)

	fmt.Println("Client Connected:", c.IP())

	//broadcastItems()

	// Read messages from the client
	for {
		msgType, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		fmt.Println("Message Type:", msgType)
		fmt.Println("Raw Message:", string(msg))
		broadcastMessage(msg)
		// Print the received message
		//fmt.Printf("Received message: %s\n", msg)

		// Send the message back to the client
		if err := c.WriteMessage(msgType, msg); err != nil {
			log.Println("Error writing message:", err)
			break
		}
	}

	initializers.WM.RemoveClient(c)
	// Connection closed
	fmt.Println("Client disconnected:", c.IP())
}

func broadcastMessage(msg []byte) error {

	initializers.WM.RLock()
	defer initializers.WM.RUnlock()

	for client := range initializers.WM.Clients {
		if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("error sending message:", err)
		}
	}

	return nil
}

func CreateItemByMigration(c *fiber.Ctx) error {

	// 1️⃣ Get Excel file
	fileHeader, err := c.FormFile("excel")
	if err != nil || fileHeader == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Missing file"})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer file.Close()

	// 2️⃣ Open Excel
	f, err := excelize.OpenReader(file)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid file"})
	}

	sheetName := f.GetSheetName(0)
	rows, _ := f.GetRows(sheetName)

	fmt.Println("Total rows:", len(rows))

	// 3️⃣ Create report
	report := excelize.NewFile()
	reportSheet := "MigrationReport"
	index, _ := report.NewSheet(reportSheet)
	report.SetActiveSheet(index)

	headers := []string{"Row", "Item Name", "Item Class", "Item Brand", "Unit of Measure", "Trade Type", "Item Tangibility Type",
		"Item Model", "Catalogue Year", "Price", "Is Stop Selling?", "Long Description", "DB_ID", "Status", "Error"}

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		report.SetCellValue(reportSheet, cell, h)
	}

	reportRow := 2

	// 4️⃣ Process each row
	for i, row := range rows {

		if i == 0 {
			continue // skip header
		}

		// 4a️⃣ Not enough columns
		if len(row) < 11 {
			itemName := ""
			itemClass := ""
			itemBrand := ""
			unitOfMeasurement := ""
			tradeType := ""
			itemTangibilityType := ""
			itemModel := ""
			catalogueYear := ""
			price := ""
			isStopSelling := ""
			ItemDescription := ""
			if len(row) > 0 {
				itemName = row[0]
				itemClass = row[1]
				itemBrand = row[2]
				unitOfMeasurement = row[3]
				tradeType = row[4]
				itemTangibilityType = row[5]
				itemModel = row[6]
				catalogueYear = row[7]
				price = row[8]
				isStopSelling = row[9]
				ItemDescription = row[10]
			}

			report.SetCellValue(reportSheet, fmt.Sprintf("A%d", reportRow), i+1)
			report.SetCellValue(reportSheet, fmt.Sprintf("B%d", reportRow), itemName)
			report.SetCellValue(reportSheet, fmt.Sprintf("C%d", reportRow), itemClass)
			report.SetCellValue(reportSheet, fmt.Sprintf("D%d", reportRow), itemBrand)
			report.SetCellValue(reportSheet, fmt.Sprintf("E%d", reportRow), unitOfMeasurement)
			report.SetCellValue(reportSheet, fmt.Sprintf("F%d", reportRow), tradeType)
			report.SetCellValue(reportSheet, fmt.Sprintf("G%d", reportRow), itemTangibilityType)
			report.SetCellValue(reportSheet, fmt.Sprintf("H%d", reportRow), itemModel)
			report.SetCellValue(reportSheet, fmt.Sprintf("I%d", reportRow), catalogueYear)
			report.SetCellValue(reportSheet, fmt.Sprintf("J%d", reportRow), price)
			report.SetCellValue(reportSheet, fmt.Sprintf("K%d", reportRow), isStopSelling)
			report.SetCellValue(reportSheet, fmt.Sprintf("L%d", reportRow), ItemDescription)
			report.SetCellValue(reportSheet, fmt.Sprintf("M%d", reportRow), 0)
			report.SetCellValue(reportSheet, fmt.Sprintf("N%d", reportRow), "Failed")
			report.SetCellValue(reportSheet, fmt.Sprintf("O%d", reportRow), "Not enough columns in Excel row")

			reportRow++
			continue
		}

		// 4b️⃣ Parse price
		price, priceErr := strconv.ParseFloat(row[8], 64)
		if priceErr != nil {
			report.SetCellValue(reportSheet, fmt.Sprintf("A%d", reportRow), i+1)
			report.SetCellValue(reportSheet, fmt.Sprintf("B%d", reportRow), row[0])
			report.SetCellValue(reportSheet, fmt.Sprintf("C%d", reportRow), row[1])
			report.SetCellValue(reportSheet, fmt.Sprintf("D%d", reportRow), row[2])
			report.SetCellValue(reportSheet, fmt.Sprintf("E%d", reportRow), row[3])
			report.SetCellValue(reportSheet, fmt.Sprintf("F%d", reportRow), row[4])
			report.SetCellValue(reportSheet, fmt.Sprintf("G%d", reportRow), row[5])
			report.SetCellValue(reportSheet, fmt.Sprintf("H%d", reportRow), row[6])
			report.SetCellValue(reportSheet, fmt.Sprintf("I%d", reportRow), row[7])
			report.SetCellValue(reportSheet, fmt.Sprintf("J%d", reportRow), row[8])
			report.SetCellValue(reportSheet, fmt.Sprintf("K%d", reportRow), row[9])
			report.SetCellValue(reportSheet, fmt.Sprintf("L%d", reportRow), row[10])
			report.SetCellValue(reportSheet, fmt.Sprintf("M%d", reportRow), 0)
			report.SetCellValue(reportSheet, fmt.Sprintf("N%d", reportRow), "Failed")
			report.SetCellValue(reportSheet, fmt.Sprintf("O%d", reportRow), "Invalid price format")
			reportRow++
			continue
		}

		// 4c️⃣ Prepare SP parameters
		conditions := map[string]interface{}{
			"ItemName":            row[0],
			"ItemClass":           row[1],
			"ItemBrand":           row[2],
			"UnitOfMeasurement":   row[3],
			"ItemTradeType":       row[4],
			"ItemTangibilityType": row[5],
			"ItemModel":           row[6],
			"CatalogueYear":       row[7],
			"Price":               price,
			"IsStopSelling":       row[9],
			"ItemDescription":     row[10],
			"ShortDescription":    "",
		}

		var result MigrationResult
		err := fetchRaw(&result, conditions)

		status := "Success"
		errMsg := result.ErrorMessage

		if err != nil {
			status = "Failed"
			errMsg = err.Error()
		}

		if result.InsertedID == 0 && errMsg != "" {
			status = "Failed"
		}

		// 4d️⃣ Write to report
		report.SetCellValue(reportSheet, fmt.Sprintf("A%d", reportRow), i+1)
		report.SetCellValue(reportSheet, fmt.Sprintf("B%d", reportRow), row[0])
		report.SetCellValue(reportSheet, fmt.Sprintf("C%d", reportRow), row[1])
		report.SetCellValue(reportSheet, fmt.Sprintf("D%d", reportRow), row[2])
		report.SetCellValue(reportSheet, fmt.Sprintf("E%d", reportRow), row[3])
		report.SetCellValue(reportSheet, fmt.Sprintf("F%d", reportRow), row[4])
		report.SetCellValue(reportSheet, fmt.Sprintf("G%d", reportRow), row[5])
		report.SetCellValue(reportSheet, fmt.Sprintf("H%d", reportRow), row[6])
		report.SetCellValue(reportSheet, fmt.Sprintf("I%d", reportRow), row[7])
		report.SetCellValue(reportSheet, fmt.Sprintf("J%d", reportRow), row[8])
		report.SetCellValue(reportSheet, fmt.Sprintf("K%d", reportRow), row[9])
		report.SetCellValue(reportSheet, fmt.Sprintf("L%d", reportRow), row[10])
		report.SetCellValue(reportSheet, fmt.Sprintf("M%d", reportRow), result.InsertedID)
		report.SetCellValue(reportSheet, fmt.Sprintf("N%d", reportRow), status)
		report.SetCellValue(reportSheet, fmt.Sprintf("O%d", reportRow), errMsg)

		reportRow++
	}

	// 5️⃣ Save and return report
	reportPath := "migration_report.xlsx"
	err = report.SaveAs(reportPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Download(reportPath)
}

// MigrationResult maps the SP output
type MigrationResult struct {
	InsertedID   int    `gorm:"column:inserted_id"`
	ErrorMessage string `gorm:"column:error_message"`
}

func fetchRaw(result *MigrationResult, conditions map[string]interface{}) error {
	query := `EXEC dbo.sp_migration_tbl_setup_item
			@ItemName = ?,
			@ItemClass = ?,
			@ItemBrand = ?,
			@UnitOfMeasurement = ?,
			@ItemTradeType = ?,
			@ItemTangibilityType = ?,
			@ItemModel = ?,
			@CatalogueYear = ?,
			@Price = ?,
			@IsStopSelling = ?,
			@ItemDescription = ?,
			@ShortDescription = ?`

	args := []interface{}{
		conditions["ItemName"],
		conditions["ItemClass"],
		conditions["ItemBrand"],
		conditions["UnitOfMeasurement"],
		conditions["ItemTradeType"],
		conditions["ItemTangibilityType"],
		conditions["ItemModel"],
		conditions["CatalogueYear"],
		conditions["Price"],
		conditions["IsStopSelling"],
		conditions["ItemDescription"],
		conditions["ShortDescription"],
	}

	return initializers.DB.Raw(query, args...).Scan(result).Error
}
