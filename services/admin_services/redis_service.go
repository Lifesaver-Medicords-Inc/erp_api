package adminservices

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/services"
)

type RedisService struct {
}

func NewRedisService() *RedisService {
	return &RedisService{}
}

func (s *RedisService) ClearAllCache(c *fiber.Ctx) (int, error) {
	if err := services.InvalidateCacheByPattern("model:*"); err != nil {
		return fiber.StatusInternalServerError, errors.New("failed getting job order data")
	}

	return 0, nil
}
