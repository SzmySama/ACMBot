package main

import (
	"github.com/YourSuzumiya/ACMBot/app/model/db"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := db.MigrateAll(); err != nil {
		logrus.Fatal(err)
	}

}
