package database

import (
	"app/src/config"
	"app/src/model"
	"app/src/utils"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(dbHost, dbName string) *gorm.DB {
	var db *gorm.DB
	var err error

	// 1. Try PostgreSQL if configured and DBHost is not empty/localhost
	if dbHost != "" && dbHost != "localhost" && dbHost != "127.0.0.1" {
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			dbHost, config.DBUser, config.DBPassword, dbName, config.DBPort,
		)
		utils.Log.Infof("Attempting to connect to PostgreSQL: %s:%d/%s", dbHost, config.DBPort, dbName)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger:                 logger.Default.LogMode(logger.Info),
			SkipDefaultTransaction: true,
			PrepareStmt:            true,
			TranslateError:         true,
		})
	}

	// 2. Fall back to SQLite local database if Postgres connection fails or isn't used
	if db == nil || err != nil {
		if err != nil {
			utils.Log.Warnf("PostgreSQL connection failed: %v. Falling back to SQLite.", err)
		} else {
			utils.Log.Info("Using SQLite local database (hris.db)")
		}

		db, err = gorm.Open(sqlite.Open("hris.db"), &gorm.Config{
			Logger:                 logger.Default.LogMode(logger.Info),
			SkipDefaultTransaction: true,
			PrepareStmt:            true,
			TranslateError:         true,
		})
		if err != nil {
			utils.Log.Fatalf("Failed to initialize SQLite database: %v", err)
		}
	}

	sqlDB, errDB := db.DB()
	if errDB != nil {
		utils.Log.Errorf("Failed to connect to database: %+v", errDB)
	} else {
		// Config connection pooling
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(60 * time.Minute)
	}

	// Run AutoMigrate
	utils.Log.Info("Running database migrations...")
	if err := db.AutoMigrate(&model.User{}, &model.Token{}); err != nil {
		utils.Log.Errorf("Failed to auto-migrate tables: %v", err)
	} else {
		utils.Log.Info("Database migrations completed successfully")
		// Seed default administrator if not present
		seedDatabase(db)
	}

	return db
}

func seedDatabase(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count == 0 {
		utils.Log.Info("Seeding database with default admin user...")
		hashedPassword, err := utils.HashPassword("admin123")
		if err != nil {
			utils.Log.Errorf("Failed to hash admin password for seeding: %v", err)
			return
		}
		admin := model.User{
			Name:          "System Admin",
			Email:         "admin@hris.com",
			Password:      hashedPassword,
			Role:          "admin",
			Department:    "IT & Security",
			Position:      "System Administrator",
			Status:        "active",
			VerifiedEmail: true,
		}
		if err := db.Create(&admin).Error; err != nil {
			utils.Log.Errorf("Failed to seed admin user: %v", err)
		} else {
			utils.Log.Info("Successfully seeded admin user: admin@hris.com / admin123")
		}
	}
}
