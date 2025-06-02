package initializers

import (
	"log"
	"os"
)

func InitLogger() {
	// Open the file in append mode; create it if it doesn't exist
	file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	// Set log output to the file
	log.SetOutput(file)
}
