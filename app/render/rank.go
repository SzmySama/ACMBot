package render

type QQUserInfo struct {
	Avatar           string
	QName            string
	CodeforcesRating uint
	RankInGroup      uint
}

type QQGroupRank struct {
	QQGroupName string
	QQUsers     []*QQUserInfo
}

func (u *QQGroupRank) ToImage() ([]byte, error) {
	return nil, nil
}
