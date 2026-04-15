package utils

import (
	"fmt"
	"strconv"
)

func WarehouseCodeGenerator(id uint) string {
	warehouseCode := fmt.Sprintf("WH#%04d", id)
	return warehouseCode
}

func AreaCodeGenerator(zone string, area string, rack string, level string, bins string) string {
	if area != "" && zone != "" {
		area = "-" + area
	}
	if rack != "" {
		rack = "-" + rack
	}
	if level != "" {
		level = "-" + level
	}

	binsCode := ""
	if b, err := strconv.Atoi(bins); err == nil && b > 0 {
		if b > 1 {
			binsCode = fmt.Sprintf("-B1 TO B%d", b)
		} else {
			binsCode = fmt.Sprintf("-B%d", b)
		}
	}

	// Concatenate all parts
	areasCode := zone + area + rack + level + binsCode
	return areasCode
}
