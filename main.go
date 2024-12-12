package main

import (
	_ "github.com/YourSuzumiya/ACMBot/app/bot/platforms/qq"
	"github.com/YourSuzumiya/ACMBot/app/model/db"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := db.MigrateAll(); err != nil {
		logrus.Fatal(err)
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
