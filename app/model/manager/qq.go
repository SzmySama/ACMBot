package manager

import (
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app/model/db"
	log "github.com/sirupsen/logrus"
	"strings"
)

func BindQQAndCodeforcesHandler(QQNumber uint, GroupNumber uint, CodeforcesHandle string) string {
	err := db.BindQQandCodeforcesHandler(QQNumber, GroupNumber, CodeforcesHandle)
	if err == nil {
		return fmt.Sprintf("成功绑定 %v -> %v", QQNumber, CodeforcesHandle)
	}
	if strings.Contains(err.Error(), "has bound by others") {
		return fmt.Sprintf("该用户名已被他人绑定！")
	}
	log.Fatal("failed to bind QQ and codeforcesHandler")
	return ""
}
