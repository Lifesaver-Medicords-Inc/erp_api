package utils

import "reflect"

func HasChanged(oldData, newData interface{}, exclude ...string) bool {
	oldValue := reflect.ValueOf(oldData)
	newValue := reflect.ValueOf(newData)

	// Dereference pointer to struct
	if oldValue.Kind() == reflect.Ptr {
		oldValue = oldValue.Elem()
	}
	if newValue.Kind() == reflect.Ptr {
		newValue = newValue.Elem()
	}

	excluded := make(map[string]bool)
	for _, ex := range exclude {
		excluded[ex] = true
	}

	// Loop all fields
	for i := 0; i < oldValue.NumField(); i++ {
		field := oldValue.Type().Field(i).Name

		if excluded[field] {
			continue
		}

		oldField := oldValue.Field(i).Interface()
		newField := newValue.Field(i).Interface()

		// Compare values
		if !reflect.DeepEqual(oldField, newField) {
			return true
		}
	}

	return false
}