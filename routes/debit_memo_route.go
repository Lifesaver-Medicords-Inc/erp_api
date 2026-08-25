package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/handlers/debit_memo_handlers"
	"github.com/pierceperado/smpc/services/debit_memo_services"
)

// DebitMemoRoutes - Debit Memo (DM#), spec §5.19/§12.6. A/P only, no
// approve endpoint - it commits entirely on save (§14.57).
func DebitMemoRoutes(app *fiber.App) {
	api := app.Group("/api/debit-memos")

	handler := debit_memo_handlers.NewDebitMemoHandler(debit_memo_services.NewDebitMemoService())

	api.Get("/", handler.GetDebitMemo)
	api.Get("/:id", handler.GetDebitMemoById)
	api.Post("/", handler.CreateDebitMemo)
}
