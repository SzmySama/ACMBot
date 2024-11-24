package db

import (
	"errors"
	"gorm.io/gorm"
)

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

func MigrateAll() (err error) {
	if err = MigrateCodeforces(); err != nil {
		return err
	}
	return nil
}
