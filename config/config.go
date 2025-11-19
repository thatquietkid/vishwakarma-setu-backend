package config

import (
	"fmt"
	"log"
	"os"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/vishwakarma-setu-backend/models"
)

var DB *gorm.DB

func ConnectDatabase() *gorm.DB {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("❌ DATABASE_URL is not set")
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}

	DB = database
	fmt.Println("✅ Connected to Database")

	err = DB.AutoMigrate(&models.Machine{})
	if err != nil {
		log.Fatalf("❌ AutoMigrate failed: %v", err)
	}

	return DB
}