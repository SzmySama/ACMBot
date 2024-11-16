package main

import (
	"github.com/YourSuzumiya/ACMBot/app/bot"
	"github.com/YourSuzumiya/ACMBot/app/model/db"
	_ "github.com/YourSuzumiya/ACMBot/app/utils/logger"
	"github.com/sirupsen/logrus"
)

func main() {
	err := db.MigrateCodeforces()
	if err != nil {
		logrus.Fatal(err)
	}
	bot.Start()
}
