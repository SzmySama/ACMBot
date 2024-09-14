package types

type QQUser struct {
	ID               int64 `gorm:"primaryKey"`
	CodeforcesHandle string
}
