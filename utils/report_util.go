package utils

import "fmt"

func DocNoGenerator(id uint) string {
	DocNo := fmt.Sprintf("RR#%04d", id)
	return DocNo
}
