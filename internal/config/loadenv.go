package config

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if err := godotenv.Load("/home/nuhmanudheen-t/Broto/2ndProject/HospitalConnect/api_gateway/.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
}
