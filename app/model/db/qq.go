package db

type QQUser struct {
	QQNumber       uint `gorm:"uniqueIndex;index:idx_qq"`
	QQGroupNumber  uint
	CodeforcesName string
}

func MigrateQQ() error {
	return db.AutoMigrate(&QQUser{})
}

func GetCodeforcesName(qqNumber uint) (string, error) {
	var codeforcesName = ""
	err := GetDBConnection().Raw(
		`SELECT codeforces_name FROM qq_users WHERE qq_number = ?`,
		qqNumber).Scan(&codeforcesName).Error
	return codeforcesName, err
}

func BindQQToCodeforcesName(user QQUser) error {
	err := GetDBConnection().Create(&user).Error
	return err
}

func UnBindQQToCodeforcesName(user QQUser) error {
	err := GetDBConnection().Where(`qq_number = ?`, user.QQNumber).Delete(user).Error
	return err
}

func ReBindQQToCodeforcesName(user QQUser) error {
	err := GetDBConnection().Model(&QQUser{}).Where(`qq_number = ?`,
		user.QQNumber).Update("codeforces_name", user.CodeforcesName).Error
	return err
}
