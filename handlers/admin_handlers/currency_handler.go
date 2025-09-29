package adminhandlers

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
	"github.com/pierceperado/smpc/utils"
)

type CurrencyHandler struct {
	CurrencyService *adminservices.CurrencyService
}

type CurrencyData struct {
	Success   bool               `json:"success"`
	Terms     string             `json:"terms"`
	Privacy   string             `json:"privacy"`
	Timestamp int64              `json:"timestamp"`
	Date      string             `json:"date"`
	Base      string             `json:"base"`
	Rates     map[string]float64 `json:"rates"`
}

type Rates struct {
	Code      string  `json:"code"`
	RateValue float64 `json:"rate_value"`
}

func NewCurrencyHandler(service *adminservices.CurrencyService) *CurrencyHandler {
	return &CurrencyHandler{
		CurrencyService: service,
	}
}

func (cu *CurrencyHandler) GetCurrencyHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")

	var data CurrencyData

	body := cu.CurrencyService.GetCurrencyAPI(idParam)
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return fiber.ErrInternalServerError
	}

	var rates []Rates
	for code, value := range data.Rates {
		rates = append(rates, Rates{
			Code:      code,
			RateValue: value,
		})
	}

	return utils.RespondSuccess(c, rates)
}
