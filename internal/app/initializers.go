package app

import (
	"log"

	"github.com/joho/godotenv"
)

func InitEnv() {
	err := godotenv.Load("main.env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}
