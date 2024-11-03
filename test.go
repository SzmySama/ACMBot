package main

import (
	"github.com/YourSuzumiya/ACMBot/app/model/db"
	"github.com/YourSuzumiya/ACMBot/app/model/fetcher"
	"github.com/YourSuzumiya/ACMBot/app/utils/logger"
	"github.com/sirupsen/logrus"
)

func main() {
	logger.Init()

	db.GetDBConnection().AutoMigrate(
		&db.CodeforcesUser{},
		&db.CodeforcesProblem{},
		&db.CodeforcesSubmission{},
		&db.CodeforcesRatingChange{},
	)

	err := fetcher.UpdateDBCodeforcesSubmissions("katago")
	if err != nil {
		logrus.Errorf("failed to update codeforces user: %v", err)
	}
}
