package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	a := App{}
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	fmt.Printf("godotenv : %s = %s \n", "DB_USERNAME", os.Getenv("DB_USERNAME"))
	fmt.Printf("godotenv : %s = %s \n", "DB_PASSWORD", os.Getenv("DB_PASSWORD"))
	fmt.Printf("godotenv : %s = %s \n", "DB_NAME", os.Getenv("DB_NAME"))

	a.initialize(
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	a.run(":8010")
}
