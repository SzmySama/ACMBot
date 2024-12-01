package db

import (
	"errors"
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app/utils/config"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	mysqldriver "github.com/go-sql-driver/mysql"
)

const (
	signalInsertLimit = 500
)

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func MigrateAll() (err error) {
	if err = MigrateCodeforces(); err != nil {
		return err
	}
	if err = MigrateQQ(); err != nil {
		return err
	}
	return nil
}

var (
	db *gorm.DB
)

func init() {
	cfg := config.GetConfig().DataBase

	if cfg.DatabaseName == "" {
		log.Fatalf("database name is empty")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DatabaseName)
	var err error
	if db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{}); err != nil {
		if !cfg.AutoCreateDB {
			log.Fatalf("failed to connect to DB: %v", err)
		}

		var mysqlErr *mysqldriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1049 {
			log.Warn(fmt.Sprintf("DataBase %s NOT exist, Creating", cfg.DatabaseName))
			// DataBase [DBName] Not Found

			/*
				1. Connect to server without select DB
				2. Create DB
				3. Use it
			*/

			// Connect to server without DB
			dsnNoDB := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
				cfg.Username, cfg.Password, cfg.Host, cfg.Port)
			db, err = gorm.Open(mysql.Open(dsnNoDB), &gorm.Config{})
			if err != nil {
				log.Fatalf("Failed to Open DataBase	 while create DB: %v", err)
			}

			// Create DB
			err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", cfg.DatabaseName)).Error
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
		}
	}
}
