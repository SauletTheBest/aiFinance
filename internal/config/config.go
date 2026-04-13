package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost string
	DBPort int
	DBUser string
	DBPassword string
	DBName string
	JWTSecret string
	ServerPort int
	OpenRouterAPIKey string
	OpenRouterModel  string
}

func Load() *Config {
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	serverPort, _ := strconv.Atoi(getEnv("SERVER_PORT", "8080"))

	return &Config{
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           dbPort,
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", ""),
		DBName:           getEnv("DB_NAME", "financial_app"),
		JWTSecret:        getEnv("JWT_SECRET", "default_secret"),
		ServerPort:       serverPort,
		OpenRouterAPIKey: getEnv("OPENROUTER_API_KEY", ""),
		OpenRouterModel:  getEnv("OPENROUTER_MODEL", "nvidia/nemotron-3-super-120b-a12b:free"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}