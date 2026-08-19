package adminservices

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

type AccessModuleService struct {
}

func NewAccessModuleService() *AccessModuleService {
	return &AccessModuleService{}
}

// GetAllAccessModulesService returns the full catalog (all apps, active rows only) so
// the Admin Access Control screen can build its App -> Module -> Submodule -> Button
// tree client-side. No pagination - 797 rows is small enough to hand over in one call,
// and the tree needs the whole set anyway to build parent/child relationships.
func (a *AccessModuleService) GetAllAccessModulesService() (*[]models.AccessModuleModel, int, error) {
	var modules = &[]models.AccessModuleModel{}

	if err := initializers.DB.
		Where("is_active = ?", true).
		Order("app_name, module, submodule, kind, button").
		Find(modules).Error; err != nil {
		return modules, fiber.StatusInternalServerError, errors.New("failed getting access modules")
	}

	return modules, fiber.StatusOK, nil
}
