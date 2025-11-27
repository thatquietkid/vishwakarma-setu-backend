package config

import (
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/vishwakarma-setu-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	DB.Migrator().DropTable(&models.Machine{})
	// fmt.Println("🔄 Running Migrations...")
	err = DB.AutoMigrate(&models.Machine{}, &models.Rental{}, &models.InspectionReport{}, &models.MaintenanceRecord{})
	if err != nil {
		log.Fatalf("❌ AutoMigrate failed: %v", err)
	}

	return DB
}
