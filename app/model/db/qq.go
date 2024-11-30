package db

import (
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
	return GetDBConnection().AutoMigrate(&QQBind{}, &QQGroup{})
}

var bindLock sync.Mutex

func BindQQToCodeforces(qqBind QQBind, group QQGroup) error {
	bindLock.Lock()
	defer bindLock.Unlock()

	return nil
}
