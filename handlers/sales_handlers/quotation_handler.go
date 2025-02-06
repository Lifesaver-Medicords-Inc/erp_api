package sales_handlers
import (
    "fmt"
    "strconv"
    "github.com/gofiber/fiber/v2"
    "github.com/pierceperado/smpc/initializers"
    "github.com/pierceperado/smpc/services/sales_services"
    "github.com/pierceperado/smpc/utils"
)
func GetSalesQuotations(c *fiber.Ctx) error {
    data, status, err := sales_services.GetSalesQuotations(nil)
    if err != nil {
        return utils.RespondError(c, status, err.Error())
    }
    return utils.RespondSuccess(c, data)
}
func GetBpis(c *fiber.Ctx) error {
    data, status, err := sales_services.GetBpis(nil)
    if err != nil {
        return utils.RespondError(c, status, err.Error())
    }
    return utils.RespondSuccess(c, data)
}
func GetItems(c *fiber.Ctx) error {
    data, status, err := sales_services.GetItems(nil)
    if err != nil {
        return utils.RespondError(c, status, err.Error())
    }
    return utils.RespondSuccess(c, data)
}
func GetBpi(c *fiber.Ctx) error {
    idParam := c.Params("id")
    idNum, err := strconv.Atoi(idParam)
    if err != nil {
        return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
    }
    data, status, err := sales_services.GetBpi(idNum)
    if err != nil {
        return utils.RespondError(c, status, err.Error())
    }
    return utils.RespondSuccess(c, data)
}
func GetItem(c *fiber.Ctx) error {
    idParam := c.Params("id")
    idNum, err := strconv.Atoi(idParam)
    if err != nil {
        fmt.Println(err)
        return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
    }
    data, status, err := sales_services.GetItem(idNum)
    if err != nil {
        fmt.Println(err)
        return utils.RespondError(c, status, err.Error())
    }
    fmt.Println(data)
    return utils.RespondSuccess(c, data)
}
func GetSalesQuotation(c *fiber.Ctx) error {
    idParam := c.Params("id")
    idNum, err := strconv.Atoi(idParam)
    if err != nil {
        return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
    }
    data, status, err := sales_services.GetSalesQuotation(idNum)
    if err != nil {
        return utils.RespondError(c, status, err.Error())
    }
    return utils.RespondSuccess(c, data)
}
func CreateSalesQuotation(c *fiber.Ctx) error {
    tx := initializers.DB.Begin()
    if tx.Error != nil {
        return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
    }
    data, status, err := sales_services.CreateSalesQuotation(c, tx)
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