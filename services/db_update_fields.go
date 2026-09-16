package services

import (
	"errors"

	"gorm.io/gorm"
)

// DbUpdateFields updates exactly the named fields of model, zero values included.
//
// DbUpdate uses UpdateColumns with a struct, and GORM skips a struct's zero values there:
// a field the user cleared, set to 0, or set back to "none" is silently left at its old
// value. Where a form sends its whole state and blank is a real answer, name the fields it
// edits here instead. Fields not named are never touched, so nothing the form does not send
// can be wiped. Names are Go field names (e.g. "LongDescription"), resolved by GORM.
func DbUpdateFields(tx *gorm.DB, model interface{}, conditions map[string]interface{}, fields ...string) error {
	if len(fields) == 0 {
		return errors.New("DbUpdateFields: no fields named")
	}

	query := tx.Model(model)
	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	if err := query.Select(fields).Updates(model).Error; err != nil {
		return err
	}

	// Same cache invalidation as DbUpdate.
	if err := InvalidateCache(GetKey(model, nil)); err != nil {
		return err
	}
	return InvalidateCacheByModel(model)
}
