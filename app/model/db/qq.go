package db

import (
	"errors"
	"github.com/YourSuzumiya/ACMBot/app/model/errs"
	"gorm.io/gorm"
	"sync"
)

type QQBind struct {
	gorm.Model
	QID              uint `gorm:"uniqueIndex"`
	CodeforcesUserID uint `gorm:"uniqueIndex"`
	QName            string
}

type QQGroup struct {
	gorm.Model
	GroupID uint `gorm:"index"`
	QID     uint
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
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err := db.Where(`q_id = ?`, qqBind.QID).First(&IDUsed).Error; err == nil {
		if err = db.Model(&QQBind{}).Where("q_id = ?", qqBind.QID).Updates(qqBind).Error; err != nil {
			return err
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err := db.Create(&qqBind).Error; err != nil {
		return err
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
