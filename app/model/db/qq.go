package db

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"sync"
)

type QQUser struct {
	gorm.Model
	QQNumber         uint     `gorm:"uniqueIndex"`
	CodeforcesHandle string   `gorm:"uniqueIndex;size:255"`
	Groups           []*Group `gorm:"many2many:user_groups"`
}

type Group struct {
	gorm.Model
	GroupID uint      `gorm:"primaryKey"`
	QQUsers []*QQUser `gorm:"many2many:user_groups"`
}

func MigrateQQ() error {
	return GetDBConnection().AutoMigrate(&QQUser{}, &Group{})
}

var bindLock sync.Mutex

func BindQQandCodeforcesHandler(QQNumber uint, GroupNumber uint, CodeforcesHandle string) error {
	bindLock.Lock()
	defer bindLock.Unlock()
	var existingUser QQUser
	if err := db.Where("codeforces_handle = ?", CodeforcesHandle).First(&existingUser).Error; err == nil {
		if existingUser.QQNumber == QQNumber {
			return nil
		}
		return fmt.Errorf("codeforces Handle %v has bound by others", CodeforcesHandle)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err := db.Where("qq_number = ?", QQNumber).First(&existingUser).Error; err == nil {
		db.Model(&QQUser{}).Where("qq_number = ?", QQNumber).Update("codeforces_handle", CodeforcesHandle)
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	user := QQUser{
		QQNumber:         QQNumber,
		CodeforcesHandle: CodeforcesHandle,
	}
	if err := db.Create(&user).Error; err != nil {
		return err
	}
	var existingGroup Group
	if err := db.Where("group_id = ?", GroupNumber).First(&existingGroup).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		existingGroup.GroupID = GroupNumber
		if err := db.Create(&existingGroup).Error; err != nil {
			return err
		}
	}
	if err := db.Model(&existingGroup).Association("QQUsers").Append(&user); err != nil {
		return err
	}
	return nil
}
