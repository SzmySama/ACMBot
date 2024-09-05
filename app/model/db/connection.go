package db

import (
	"fmt"

	"github.com/SzmySama/ACMBot/app/types"
	"github.com/SzmySama/ACMBot/app/utils/config"
	mysqldriver "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func GetDBConnection() *gorm.DB {
	return db
}

func InitDB(autoCreate bool) error {
	cfg := config.GetConfig().DataBase
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DatabaseName)
	var err error
	if db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{}); err != nil {
		if !autoCreate {
			return fmt.Errorf("failed to connect to DB: %w", err)
		}

		if mysqlErr, ok := err.(*mysqldriver.MySQLError); ok && mysqlErr.Number == 1049 {
			log.Warn(fmt.Sprintf("DataBase %s NOT exist, Creating", cfg.DatabaseName))
			// DataBase [DBName] Not Found

			/*
				1. Connect to server without select DB
				2. Create DB
				3. Use it
			*/

			// Connect to server without DB
			dsnNoDB := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
				cfg.Username, cfg.Password, cfg.Host, cfg.Port)
			db, err = gorm.Open(mysql.Open(dsnNoDB), &gorm.Config{})
			if err != nil {
				log.Fatalf("Failed to Open DataBase	 while create DB: %v", err)
			}

			// Create DB
			SQLCreatDB := fmt.Sprintf(`
					CREATE DATABASE IF NOT EXISTS %s
				`, cfg.DatabaseName)
			err = db.Exec(SQLCreatDB).Error
			if err != nil {
				log.Fatalf("Failed to create DataBase: %v", err)
			}
			log.Infof("Create DB %s Successfully", cfg.DatabaseName)

			// Use it
			SQLUseDB := fmt.Sprintf(`
					USE %s
				`, cfg.DatabaseName)
			err = db.Exec(SQLUseDB).Error
			if err != nil {
				log.Fatalf("Failed to use database %v: %v", cfg.DatabaseName, err)
			}
		} else {
			log.Fatalf("Failed to Open DataBase:%v", err)
		}
	}
	return nil
}

func Migrate() error {
	err := db.AutoMigrate(&types.User{}, &types.Problem{}, &types.Submission{}, &types.RatingChange{})
	if err != nil {
		return fmt.Errorf("failed to migrate DB: %w", err)
	}
	return nil
}
