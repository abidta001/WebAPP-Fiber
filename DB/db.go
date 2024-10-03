package db

import (
	"fmt"
	"log"
	models "myapp/Models"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Db *gorm.DB

func InitDatabase() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env: ", err)
	}

	// Print the DSN for debugging purposes
	dsn := os.Getenv("dsn")
	fmt.Println("DSN: ", dsn)

	Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
		return
	}

	err = Db.AutoMigrate(&models.User{})
	if err != nil {
		fmt.Println("Error in automigrating User: ", err)
	}

	err = Db.AutoMigrate(&models.Admin{})
	if err != nil {
		fmt.Println("Error in automigrating Admin: ", err)
	}

	fmt.Println("Database connected successfully")
}
