package main

import (
	"github.com/SzmySama/ACMBot/app/bot"
	"github.com/SzmySama/ACMBot/app/utils/logger"
)

func main() {
	logger.Init()
	bot.Start()
}
