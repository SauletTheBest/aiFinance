package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/app"
	"github.com/SauletTheBest/BackendFinancialApplication/internal/config"
)

func main() {
	// 1. Load .env file (if it exists)
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found")
	}

	// 2. Load all configuration variables
	cfg := config.Load()

	// 3. Create the Application (we inject the config here!)
	application := app.NewApp(cfg)

	// 4. Run the Application (this function blocks until you press Ctrl+C)
	application.Run()
}
