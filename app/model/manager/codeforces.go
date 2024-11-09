package manager

import (
	"encoding/json"
	"errors"
	"github.com/YourSuzumiya/ACMBot/app/model/cache"
	"github.com/YourSuzumiya/ACMBot/app/model/db"
	"github.com/YourSuzumiya/ACMBot/app/model/fetcher"
	"github.com/YourSuzumiya/ACMBot/app/model/render"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"sync"
	"time"
)

var (
	updatingUser sync.Map
)

/*
CodeforcesUser
从数据库中加载建议使用LoadFromDB函数
如需手动操作
必须preload所有RatingChanges
必须preload最后一条Submission
才能执行相关函数
*/
type CodeforcesUser struct {
	DBUser    db.CodeforcesUser
	MaxRating uint
}

func (u *CodeforcesUser) LoadFromDB(handle string) error {
	return mdb.
		Preload("Submissions", func(db *gorm.DB) *gorm.DB { return db.Order("at DESC").Limit(1) }).
		Preload("RatingChanges").Where("handle = ?", handle).First(&u.DBUser).Error
}

func (u *CodeforcesUser) FromFetcherUserInfo(user fetcher.CodeforcesUser) error {
	u.DBUser.Handle = user.Handle
	u.DBUser.Avatar = user.Avatar
	u.DBUser.Rating = user.Rating
	u.DBUser.FriendCount = user.FriendCount
	u.DBUser.CreatedAt = user.CreatedAt
	return nil
}

func (u *CodeforcesUser) FromFetcherRatingChanges(changes []fetcher.CodeforcesRatingChange) error {
	lastRatingChangeInDB := time.Unix(0, 0)
	if len(u.DBUser.RatingChanges) > 0 {
		lastRatingChangeInDB = u.DBUser.RatingChanges[len(u.DBUser.RatingChanges)-1].At
	}

	firstNewRatingChangeIdx := -1

	for idx, change := range changes {
		if change.At.After(lastRatingChangeInDB) {
			firstNewRatingChangeIdx = idx
			break
		}
	}

	if firstNewRatingChangeIdx == -1 {
		return nil
	}

	for _, change := range changes[firstNewRatingChangeIdx:] {
		u.DBUser.RatingChanges = append(u.DBUser.RatingChanges, db.CodeforcesRatingChange{
			CodeforcesUserID: u.DBUser.ID,
			At:               change.At,
			NewRating:        change.NewRating,
		})
	}
	return nil
}

func (u *CodeforcesUser) FromFetcherSubmissions(submissions []fetcher.CodeforcesSubmission) error {
	lastSubmissionInDB := time.Unix(0, 0)
	if len(u.DBUser.Submissions) > 0 {
		lastSubmissionInDB = u.DBUser.Submissions[len(u.DBUser.Submissions)-1].At
	}

	lastOldSubmissionIdx := len(submissions)

	for i := len(submissions) - 1; i >= 0; i-- {
		if !submissions[i].At.After(lastSubmissionInDB) {
			lastOldSubmissionIdx = i
		} else {
			break
		}
	}

	problemID2submission := make(map[string][]db.CodeforcesSubmission)
	problemID2problem := make(map[string]db.CodeforcesProblem)
	newSubmissions := make([]db.CodeforcesSubmission, 0, len(submissions[:lastOldSubmissionIdx]))

	for _, submission := range submissions[:lastOldSubmissionIdx] {
		id := submission.Problem.ID()
		dbSubmission := db.CodeforcesSubmission{
			CodeforcesUserID:    u.DBUser.ID,
			CodeforcesProblemID: id,
			At:                  submission.At,
			Status:              submission.Status,
		}

		newSubmissions = append(newSubmissions, dbSubmission)
		problemID2submission[id] = append(problemID2submission[id], dbSubmission)
		problemID2problem[id] = db.CodeforcesProblem{
			ID:     id,
			Rating: submission.Problem.Rating,
		}
	}

	s := make([]db.CodeforcesProblem, 0, len(problemID2problem))
	for id, _ := range problemID2submission {
		s = append(s, db.CodeforcesProblem{
			ID:     id,
			Rating: problemID2problem[id].Rating,
		})
	}

	if len(s) > 0 {
		if result := mdb.Save(s); result.Error != nil {
			return result.Error
		}
	}

	u.DBUser.Submissions = append(u.DBUser.Submissions, newSubmissions...)
	return nil
}

func (u *CodeforcesUser) SaveToDB() error {
	//每次最多插入5000条submission
	const mx = 5000
	var err error
	var submissions []db.CodeforcesSubmission
	if len(u.DBUser.Submissions) > mx {
		submissions = u.DBUser.Submissions
		u.DBUser.Submissions = []db.CodeforcesSubmission{}
	}
	if err = mdb.Save(&u.DBUser).Error; err != nil {
		return err
	}
	if submissions != nil {
		for i, _ := range submissions {
			submissions[i].CodeforcesUserID = u.DBUser.ID
		}

		for i := 0; i < len(submissions); i += mx {
			if err = mdb.Save(submissions[i:min(len(submissions), i+mx)]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *CodeforcesUser) ToRenderUser() *render.CodeforcesUser {
	return &render.CodeforcesUser{
		Handle:    u.DBUser.Handle,
		Avatar:    u.DBUser.Avatar,
		Rating:    u.DBUser.Rating,
		Solved:    u.DBUser.Solved,
		FriendOf:  u.DBUser.FriendCount,
		CreatedAt: u.DBUser.CreatedAt,
		Level:     render.ConvertRatingToLevel(u.DBUser.Rating),
	}
}

func (u *CodeforcesUser) ToRenderRatingChanges() *render.CodeforcesRatingChanges {
	ratingChanges := make([]render.CodeforcesRatingChange, len(u.DBUser.RatingChanges))
	for i, change := range u.DBUser.RatingChanges {
		ratingChanges[i] = render.CodeforcesRatingChange{
			At:        change.At,
			NewRating: change.NewRating,
		}
	}
	return &render.CodeforcesRatingChanges{
		Data:   ratingChanges,
		Handle: u.DBUser.Handle,
	}
}

func GetUpdatedCodeforcesUser(handle string) (*CodeforcesUser, error) {
	const normalSubmissionFetchNum = 500
	const newUserSubmissionFetchNum = 10000

	v, _ := updatingUser.LoadOrStore(handle, &sync.Mutex{})
	lock := v.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	cacheUser, err := cache.GetCodeforcesUser(handle)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	result := CodeforcesUser{}
	if cacheUser != "" && json.Unmarshal([]byte(cacheUser), &result) == nil {
		return &result, nil
	}

	isNewUser := false
	if err = result.LoadFromDB(handle); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		isNewUser = true
	}

	userInfo, err := fetcher.FetchCodeforcesUsersInfo([]string{handle}, false)
	if err != nil {
		return nil, err
	}

	ratingChanges, err := fetcher.FetchCodeforcesUserRatingChanges(handle)
	if err != nil {
		return nil, err
	}

	fetchNum := normalSubmissionFetchNum
	if isNewUser {
		fetchNum = newUserSubmissionFetchNum
	}

	lastSubmissionInDB := time.Unix(0, 0)
	if len(result.DBUser.Submissions) > 0 {
		lastSubmissionInDB = result.DBUser.Submissions[len(result.DBUser.Submissions)-1].At
	}

	submissions := make([]fetcher.CodeforcesSubmission, 0)
	count := 1
	flag := true
	for flag {
		correctSubmissions, err := fetcher.FetchCodeforcesUserSubmissions(handle, count, fetchNum)
		if err != nil {
			return nil, err
		}
		if len(*correctSubmissions) == 0 {
			break
		}
		flag = (*correctSubmissions)[0].At.Before(lastSubmissionInDB)
		count += fetchNum
		submissions = append(submissions, *correctSubmissions...)
	}

	if err = result.FromFetcherRatingChanges(*ratingChanges); err != nil {
		return nil, err
	}

	if err = result.FromFetcherSubmissions(submissions); err != nil {
		return nil, err
	}

	if err = result.FromFetcherUserInfo((*userInfo)[0]); err != nil {
		return nil, err
	}

	if isNewUser {
		solved := make(map[string]byte)
		for _, submission := range submissions {
			if submission.Status == string(db.CodeforcesSubmissionStatusOk) {
				solved[submission.Problem.ID()] = 1
			}
		}
		result.DBUser.Solved = uint(len(solved))
	} else {
		if result := db.GetDBConnection().Raw(`
		SELECT COUNT(DISTINCT codeforces_problem_id) 
		FROM codeforces_submissions 
		WHERE codeforces_user_id = ? AND status = ?`,
			result.DBUser.ID, db.CodeforcesSubmissionStatusOk).Scan(&result.DBUser.Solved); result.Error != nil {
			return nil, result.Error
		}
	}

	if err = result.SaveToDB(); err != nil {
		return nil, err
	}

	// Do Not show submission to outside
	result.DBUser.Submissions = nil

	for _, change := range *ratingChanges {
		result.MaxRating = max(result.MaxRating, uint(change.NewRating))
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	if err = cache.SetCodeforcesUser(handle, data, 4*time.Hour); err != nil {
		return nil, err
	}

	return &result, nil
}
