package db

import (
	"errors"
	"github.com/YourSuzumiya/ACMBot/app/errs"
	"gorm.io/gorm"
	"sync"
)

type QQBind struct {
	gorm.Model
	QID              uint64 `gorm:"uniqueIndex"`
	CodeforcesUserID uint   `gorm:"uniqueIndex"`
	QName            string
}

type QQGroup struct {
	gorm.Model
	GroupID uint64 `gorm:"index"`
	QID     uint64
}

type QQUser struct {
	QID              uint64
	CodeforcesRating uint
	QName            string
}

type QQGroupRank struct {
	GroupID uint64
	QQUsers []*QQUser
}

func MigrateQQ() error {
	return db.AutoMigrate(&QQBind{}, &QQGroup{})
}

var (
	bindCfID sync.Map
	bindQID  sync.Map
)

func BindQQToCodeforces(qqBind QQBind, group QQGroup) error {
	v, _ := bindCfID.LoadOrStore(qqBind.CodeforcesUserID, &sync.Mutex{})
	u, _ := bindQID.LoadOrStore(qqBind.QID, &sync.Mutex{})
	cfIDLock := v.(*sync.Mutex)
	qIDLock := u.(*sync.Mutex)
	cfIDLock.Lock()
	qIDLock.Lock()
	defer cfIDLock.Unlock()
	defer qIDLock.Unlock()
	var IDUsed QQBind
	if err := db.Where(`codeforces_user_id = ?`, qqBind.CodeforcesUserID).First(&IDUsed).Error; err == nil {
		if IDUsed.QID != qqBind.QID {
			return errs.ErrHandleHasBindByOthers
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	} else {
		if err = db.Where(`q_id = ?`, qqBind.QID).First(&IDUsed).Error; err == nil {
			if err = db.Model(&QQBind{}).Where("q_id = ?", qqBind.QID).Updates(qqBind).Error; err != nil {
				return err
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err = db.Create(&qqBind).Error; err != nil {
			return err
		}
	}
	var existGroup QQGroup
	if err := db.Where(`group_id = ? and q_id = ?`, group.GroupID, group.QID).First(&existGroup).Error; err == nil {
		return nil
	}
	if err := db.Create(&group).Error; err != nil {
		return err
	}
	return nil
}

func CheckBind(QID uint) (bool, error) {
	if err := db.Where("q_id = ?", QID).First(&QQBind{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func GetQQGroupRank(QQGroupNumber uint64) (*QQGroupRank, error) {
	var QIDs []uint64
	if err := db.Model(&QQGroup{}).Where("group_id = ?", QQGroupNumber).Pluck("q_id", &QIDs).Error; err != nil {
		return nil, err
	}
	var qqGroupRank = QQGroupRank{
		GroupID: QQGroupNumber,
		QQUsers: make([]*QQUser, 0),
	}
	for _, QID := range QIDs {
		var qqBind QQBind
		if err := db.Where("q_id = ?", QID).First(&qqBind).Error; err != nil {
			return nil, err
		}
		var cfRating uint
		if err := db.Model(&CodeforcesUser{}).Select("rating").Where("id = ?", qqBind.CodeforcesUserID).First(&cfRating).Error; err != nil {
			return nil, err
		}
		var qqUser = QQUser{
			QID:              QID,
			QName:            qqBind.QName,
			CodeforcesRating: cfRating,
		}
		qqGroupRank.QQUsers = append(qqGroupRank.QQUsers, &qqUser)
	}
	return &qqGroupRank, nil
}
