package manager

import (
	"errors"
	"github.com/YourSuzumiya/ACMBot/app/model/db"
	"github.com/YourSuzumiya/ACMBot/app/model/errs"
	"github.com/YourSuzumiya/ACMBot/app/model/fetcher"
	log "github.com/sirupsen/logrus"
)

type QQBind struct {
	QQGroupID        uint
	QQName           string
	QID              uint
	CodeforcesHandle string
}

func BindQQAndCodeforcesHandler(qqBind QQBind) error {
	var err error
	var user *fetcher.CodeforcesUser
	if user, err = fetcher.FetchCodeforcesUserInfo(qqBind.CodeforcesHandle, false); err != nil {
		if errors.Is(err, errs.ErrHandleNotFound) {
			return err
		}
		log.Fatalf("fetch failed %v", err)
		return err
	}
	if user.Organization != "ACMBot" {
		return errs.ErrOrganizationUnmatched
	}
	if _, err = GetUpdatedCodeforcesUser(qqBind.CodeforcesHandle); err != nil {
		return err
	}
	var userID uint
	if userID, err = db.GetCodeforcesUserID(qqBind.CodeforcesHandle); err != nil {
		log.Fatalf("get code forces user id %v", err)
		return err
	}
	var bind = db.QQBind{
		QID:              qqBind.QID,
		CodeforcesUserID: userID,
		QName:            qqBind.QQName,
	}
	var group = db.QQGroup{
		GroupID: qqBind.QQGroupID,
		QID:     qqBind.QID,
	}
	if err = db.BindQQToCodeforces(bind, group); err != nil {
		if !errors.Is(err, errs.ErrHandleHasBindByOthers) {
			log.Fatalf("bind QQ in db failed %v", err)
		}
		return err
	}
	return nil
}
