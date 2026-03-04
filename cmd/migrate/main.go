package main

import (
	"log"
	"os"

	"github.com/birabittoh/rainbbit/src"
	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "data/data.sqlite"
	}

	pgDSN := os.Getenv("DATABASE_URL")
	if pgDSN == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	log.Printf("Connecting to SQLite: %s", sqlitePath)
	sqliteDB, err := gorm.Open(sqlite.Open(sqlitePath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to SQLite: %v", err)
	}

	log.Println("Connecting to PostgreSQL...")
	pgDB, err := gorm.Open(postgres.Open(pgDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	log.Println("Migrating schema to PostgreSQL...")
	err = pgDB.AutoMigrate(&src.Record{})
	if err != nil {
		log.Fatalf("Failed to migrate PostgreSQL schema: %v", err)
	}

	var records []src.Record
	log.Println("Fetching records from SQLite...")
	err = sqliteDB.Find(&records).Error
	if err != nil {
		log.Fatalf("Failed to fetch records from SQLite: %v", err)
	}

	log.Printf("Found %d records. Starting migration...", len(records))

	if len(records) > 0 {
		// Batch insert to PostgreSQL
		err = pgDB.CreateInBatches(records, 100).Error
		if err != nil {
			log.Fatalf("Failed to insert records into PostgreSQL: %v", err)
		}
	}

	log.Println("Migration completed successfully!")
}
