package manager

import (
	"errors"
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app/errs"
	"github.com/YourSuzumiya/ACMBot/app/fetcher"
	"github.com/YourSuzumiya/ACMBot/app/model/db"
	"github.com/YourSuzumiya/ACMBot/app/render"
	log "github.com/sirupsen/logrus"
	"sort"
)

type QQBind struct {
	QQGroupID        uint64
	QQName           string
	QID              uint64
	CodeforcesHandle string
}

type QQGroup struct {
	QQGroupName string
	QQGroupID   uint64
}

func BindQQAndCodeforcesHandler(qqBind QQBind) error {
	var err error
	var user *fetcher.CodeforcesUser
	if user, err = fetcher.FetchCodeforcesUserInfo(qqBind.CodeforcesHandle, false); err != nil {
		if errors.Is(err, errs.ErrHandleNotFound{}) {
			return err
		}
		log.Errorf("fetch failed %v", err)
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
		log.Errorf("get code forces user id %v", err)
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
			log.Errorf("bind CallerID in db failed %v", err)
		}
		return err
	}
	return nil
}

func GetGroupRank(qqGroup QQGroup) (*render.QQGroupRank, error) {
	var rank *db.QQGroupRank
	var err error
	if rank, err = db.GetQQGroupRank(qqGroup.QQGroupID); err != nil {
		log.Errorf("get group rank failed %v", err)
		return nil, err
	}
	sort.Slice(rank.QQUsers, func(i, j int) bool {
		return rank.QQUsers[i].CodeforcesRating > rank.QQUsers[j].CodeforcesRating
	})
	var qqGroupRank = render.QQGroupRank{
		QQGroupName: qqGroup.QQGroupName,
		QQUsers:     make([]*render.QQUserInfo, 0),
	}
	for index, user := range rank.QQUsers {
		var QQUser = render.QQUserInfo{
			QName:            user.QName,
			CodeforcesRating: user.CodeforcesRating,
			RankInGroup:      uint(index) + 1,
			Avatar:           "https://q1.qlogo.cn/g?b=qq&nk=" + fmt.Sprintf("%d", user.QID) + "&s=4",
		}
		qqGroupRank.QQUsers = append(qqGroupRank.QQUsers, &QQUser)
	}
	return &qqGroupRank, nil
}
