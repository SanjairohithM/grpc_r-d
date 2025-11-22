// +build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Run migrations standalone
func main() {
	// Load DATABASE_URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set. Please create .env file first.")
	}

	fmt.Println("📊 Connecting to database...")
	
	// Connect to database
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	fmt.Println("✅ Connected to database")
	fmt.Println("🔄 Running migrations...")
	
	// Run migrations
	if err := db.AutoMigrate(&User{}, &Greeting{}); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	fmt.Println("✅ Migrations completed successfully!")
	fmt.Println("")
	fmt.Println("Created/Updated tables:")
	fmt.Println("  ✓ users")
	fmt.Println("  ✓ greetings")
	fmt.Println("")
	fmt.Println("🎉 Database is ready!")
}

