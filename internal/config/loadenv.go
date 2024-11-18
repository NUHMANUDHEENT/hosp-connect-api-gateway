package config

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if err := godotenv.Load("./cmd/.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
}
