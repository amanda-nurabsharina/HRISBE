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

var DB *gorm.DB

func Connect(dbHost, dbName string) *gorm.DB {
	var db *gorm.DB
	var err error

	// 1. Try PostgreSQL using DBConfig if Host is configured
	if config.DBConfig.Host != "" {
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
			config.DBConfig.Host, config.DBConfig.User, config.DBConfig.Password, config.DBConfig.DBName, config.DBConfig.Port, config.DBConfig.SSLMode,
		)
		utils.Log.Infof("Attempting to connect to PostgreSQL: %s:%s/%s", config.DBConfig.Host, config.DBConfig.Port, config.DBConfig.DBName)
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
	if err := db.AutoMigrate(&model.User{}, &model.Token{}, &model.Role{}); err != nil {
		utils.Log.Errorf("Failed to auto-migrate tables: %v", err)
	} else {
		utils.Log.Info("Database migrations completed successfully")
		// Seed default administrator if not present
		seedDatabase(db)
	}

	DB = db
	return db
}

func seedDatabase(db *gorm.DB) {
	// 1. Seed Roles
	utils.Log.Info("Seeding roles...")
	for _, roleCfg := range config.DefaultRoles {
		var count int64
		db.Model(&model.Role{}).Where("name = ?", roleCfg.Name).Count(&count)
		if count == 0 {
			newRole := model.Role{
				Name:            roleCfg.Name,
				DisplayName:     roleCfg.DisplayName,
				Description:     roleCfg.Description,
				AccessibleMenus: model.StringArray(roleCfg.AccessibleMenus),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			if err := db.Create(&newRole).Error; err != nil {
				utils.Log.Errorf("Failed to seed role %s: %v", roleCfg.Name, err)
			} else {
				utils.Log.Infof("Seeded role: %s", roleCfg.Name)
			}
		}
	}

	// 2. Seed Users from config (superadmin, etc.)
	utils.Log.Info("Seeding default users...")
	for _, userCfg := range config.DefaultUsers {
		var count int64
		db.Model(&model.User{}).Where("email = ?", userCfg.Email).Count(&count)
		if count == 0 {
			hashedPassword, err := utils.HashPassword(userCfg.Password)
			if err != nil {
				utils.Log.Errorf("Failed to hash password for user %s: %v", userCfg.Username, err)
				continue
			}
			newUser := model.User{
				Name:          userCfg.FirstName + " " + userCfg.LastName,
				Email:         userCfg.Email,
				Password:      hashedPassword,
				Role:          userCfg.Role,
				Department:    "Executive",
				Position:      "Super Administrator",
				Status:        "active",
				VerifiedEmail: true,
			}
			if err := db.Create(&newUser).Error; err != nil {
				utils.Log.Errorf("Failed to seed user %s: %v", userCfg.Email, err)
			} else {
				utils.Log.Infof("Successfully seeded user: %s (role: %s)", userCfg.Email, userCfg.Role)
			}
		}
	}

	// 3. Seed legacy admin if no user exists at all
	var userCount int64
	db.Model(&model.User{}).Count(&userCount)
	if userCount == 0 || userCount == 1 { // if only super_admin exists or 0 users
		var legacyCount int64
		db.Model(&model.User{}).Where("email = ?", "admin@hris.com").Count(&legacyCount)
		if legacyCount == 0 {
			utils.Log.Info("Seeding database with legacy admin user...")
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
				utils.Log.Info("Successfully seeded legacy admin user: admin@hris.com / admin123")
			}
		}
	}
}
