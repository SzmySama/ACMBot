package main

import (
	"github.com/SzmySama/ACMBot/app/bot"
	"github.com/SzmySama/ACMBot/app/model/db"
	"github.com/SzmySama/ACMBot/app/utils/logger"
	"github.com/sirupsen/logrus"
)

func main() {
	logger.Init()
	err := db.InitDB(true)
	if err != nil {
		logrus.Errorf("failed to init db: %v", err)
		return
	}
	err = db.Migrate()
	if err != nil {
		logrus.Errorf("failed to migrate db: %v", err)
		return
	}
	bot.Start()
}
