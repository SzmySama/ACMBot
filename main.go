package main

import (
	"github.com/SzmySama/ACMBot/app/bot"
	"github.com/SzmySama/ACMBot/app/model/db"
	"github.com/SzmySama/ACMBot/app/utils/logger"
)

func main() {
	logger.Init()
	db.InitDB(true)
	db.Migrate()
	bot.Start()
}

// func main() {
// 	logger.Init()

// }
