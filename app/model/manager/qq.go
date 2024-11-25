package manager

import (
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app/model/db"
	"github.com/YourSuzumiya/ACMBot/app/model/fetcher"
	log "github.com/sirupsen/logrus"
	"strings"
)

func BindQQAndCodeforcesHandler(QQNumber uint, GroupNumber uint, CodeforcesHandle string) error {
	userInfo, err := fetcher.FetchCodeforcesUsersInfo([]string{CodeforcesHandle}, false)
	if err != nil {
		return err
	}
	organization := (*userInfo)[0].Organization
	if organization != "ACMBot" {
		return fmt.Errorf("该cf账号不在ACMBot里，请先修改账号所在的organization再绑定哦")
	}
	if err = db.BindQQandCodeforcesHandler(QQNumber, GroupNumber, CodeforcesHandle); err != nil {
		if strings.Contains(err.Error(), "has bound by others") {
			return fmt.Errorf(CodeforcesHandle + "已被他人绑定！")
		} else {
			log.Fatal("failed to bind QQ and codeforcesHandler")
			return fmt.Errorf("failed to bind QQ and codeforcesHandler")
		}
	}
	if _, err = GetUpdatedCodeforcesUser(CodeforcesHandle); err != nil {
		return err
	}
	return nil
}
