package main

import (
	"github.com/YourSuzumiya/ACMBot/app/bot"
	"github.com/YourSuzumiya/ACMBot/app/model/db"
	_ "github.com/YourSuzumiya/ACMBot/app/utils/logger"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := db.MigrateCodeforces(); err != nil {
		logrus.Fatal(err)
	}
	if err := db.MigrateQQ(); err != nil {
		logrus.Fatal(err)
	}
	bot.Start()
}
