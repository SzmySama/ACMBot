package manager

import (
	"encoding/json"
	"github.com/YourSuzumiya/ACMBot/app/fetcher"
	"github.com/YourSuzumiya/ACMBot/app/render"
	"sort"
	"sync"
	"time"

	"github.com/YourSuzumiya/ACMBot/app/model/cache"
	"github.com/YourSuzumiya/ACMBot/app/model/db"
)

var (
	updatingUser sync.Map
)

/*
 0 -> 0~800
 800 -> 800~1200
 1200 -> 1200~1600
...
*/

type SolvedData struct {
	RatingRange int
	Count       int
}

/*
CodeforcesUser
必须preload所有RatingChanges和最后一条Submission
才能执行相关函数
*/
type CodeforcesUser struct {
	DBUser         db.CodeforcesUser
	SolvedProblems []SolvedData
	SolvedCount    int
}

func (u *CodeforcesUser) ToRenderProfileV1() *render.CodeforcesUser {
	return &render.CodeforcesUser{
		Handle:   u.DBUser.Handle,
		Avatar:   u.DBUser.Avatar,
		Rating:   u.DBUser.Rating,
		Solved:   u.SolvedCount,
		FriendOf: u.DBUser.FriendCount,
		Level:    render.ConvertRatingToLevel(u.DBUser.Rating),
	}
}

func (u *CodeforcesUser) ToRenderRatingChanges() *render.CodeforcesRatingChanges {
	ratingChanges := make([]render.CodeforcesRatingChange, len(u.DBUser.RatingChanges))
	for i, change := range u.DBUser.RatingChanges {
		ratingChanges[i] = render.CodeforcesRatingChange{
			At:        change.At.Unix(),
			NewRating: change.NewRating,
		}
	}
	return &render.CodeforcesRatingChanges{
		Data:   ratingChanges,
		Handle: u.DBUser.Handle,
	}
}

func (u *CodeforcesUser) ToRenderProfileV2() *render.CodeforcesUserProfile {
	result := make([]render.CodeforcesUserSolvedData, 4)

	result[0].Range = "800+"
	result[1].Range = "1400+"
	result[2].Range = "2000+"
	result[3].Range = "2600+"

	for _, problem := range u.SolvedProblems {
		switch {
		case problem.RatingRange < 800:
		case problem.RatingRange < 1400:
			result[0].Count += problem.Count
		case problem.RatingRange < 2000:
			result[1].Count += problem.Count
		case problem.RatingRange < 2600:
			result[2].Count += problem.Count
		default:
			result[3].Count += problem.Count
		}
	}

	for k, v := range result {
		result[k].Percent = float32(v.Count) / float32(u.SolvedCount) * 100
	}

	return &render.CodeforcesUserProfile{
		Avatar:     u.DBUser.Avatar,
		Handle:     u.DBUser.Handle,
		MaxRating:  u.DBUser.MaxRating,
		FriendOf:   u.DBUser.FriendCount,
		Rating:     u.DBUser.Rating,
		Solved:     u.SolvedCount,
		Level:      render.ConvertRatingToLevel(u.DBUser.Rating),
		SolvedData: result,
	}
}

func (u *CodeforcesUser) fromFetcherUserInfo(user *fetcher.CodeforcesUser) error {
	u.DBUser.Handle = user.Handle
	u.DBUser.Avatar = user.Avatar
	u.DBUser.Rating = user.Rating
	u.DBUser.FriendCount = user.FriendCount
	u.DBUser.MaxRating = user.MaxRating
	u.DBUser.CreatedAt = time.Unix(user.CreatedAt, 0)
	return nil
}

func (u *CodeforcesUser) fromFetcherRatingChanges(changes []fetcher.CodeforcesRatingChange) error {
	for _, change := range changes {
		u.DBUser.RatingChanges = append(u.DBUser.RatingChanges, db.CodeforcesRatingChange{
			CodeforcesUserID: u.DBUser.ID,
			At:               time.Unix(change.At, 0),
			NewRating:        change.NewRating,
		})
	}
	return nil
}

func (u *CodeforcesUser) fromFetcherSubmissions(submissions []fetcher.CodeforcesSubmission) error {
	problemID2submission := make(map[string][]db.CodeforcesSubmission)
	problemID2problem := make(map[string]db.CodeforcesProblem)
	newSubmissions := make([]db.CodeforcesSubmission, 0, len(submissions))

	for _, submission := range submissions {
		id := submission.Problem.ID()
		dbSubmission := db.CodeforcesSubmission{
			CodeforcesUserID:    u.DBUser.ID,
			CodeforcesProblemID: id,
			At:                  time.Unix(submission.At, 0),
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
	for id := range problemID2submission {
		s = append(s, db.CodeforcesProblem{
			ID:     id,
			Rating: problemID2problem[id].Rating,
		})
	}

	if len(s) > 0 {
		if err := db.SaveCodeforcesProblems(s); err != nil {
			return err
		}
	}

	u.DBUser.Submissions = append(u.DBUser.Submissions, newSubmissions...)
	return nil
}

func (u *CodeforcesUser) loadFromDB(handle string) error {
	if user, err := db.LoadCodeforcesUserByHandle(handle); err != nil {
		return err
	} else {
		u.DBUser = *user
		return nil
	}
}

func (u *CodeforcesUser) saveUser2DB() error {
	user := u.DBUser
	user.RatingChanges = nil
	user.Submissions = nil
	err := db.SaveCodeforcesUser(&user)
	if err != nil {
		return err
	}
	u.DBUser.ID = user.ID
	return nil
}

func (u *CodeforcesUser) saveFetcherSubmissions2DB(submissions []fetcher.CodeforcesSubmission) error {
	dbSubmissions := make([]db.CodeforcesSubmission, 0, len(submissions))
	for _, submission := range submissions {
		dbSubmissions = append(dbSubmissions, db.CodeforcesSubmission{
			CodeforcesUserID:    u.DBUser.ID,
			CodeforcesProblemID: submission.Problem.ID(),
			At:                  time.Unix(submission.At, 0),
			Status:              submission.Status,
		})
	}
	return db.SaveCodeforcesSubmissions(dbSubmissions)
}

func (u *CodeforcesUser) saveFetcherRatingChanges2DB(ratingChanges []fetcher.CodeforcesRatingChange) error {
	dbRatingChanges := make([]db.CodeforcesRatingChange, 0, len(ratingChanges))
	for _, ratingChange := range ratingChanges {
		dbRatingChanges = append(dbRatingChanges, db.CodeforcesRatingChange{
			CodeforcesUserID: u.DBUser.ID,
			At:               time.Unix(ratingChange.At, 0),
			NewRating:        ratingChange.NewRating,
		})
	}
	return db.SaveCodeforcesRatingChanges(dbRatingChanges)
}

func (u *CodeforcesUser) loadFromCache(handle string) (err error) {
	var data string
	if data, err = cache.GetCodeforcesUser(handle); err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), &u)
}

func (u *CodeforcesUser) saveToCache() (err error) {
	var data []byte
	if data, err = json.Marshal(u); err != nil {
		return err
	}
	if err = cache.SetCodeforcesUser(u.DBUser.Handle, data, 4*time.Hour); err != nil {
		return err
	}
	return nil
}

// cruDB create/read & update user data in DB
func (u *CodeforcesUser) cruDBUser(handle string) (err error) {
	isNewUser := false
	if err = u.loadFromDB(handle); err != nil {
		if !db.IsNotFound(err) {
			return err
		}
		isNewUser = true
	}

	userInfo, err := fetcher.FetchCodeforcesUserInfo(handle, false)
	if err != nil {
		return err
	}

	lastRatingChangeAt := time.Unix(0, 0)
	if !isNewUser {
		lastRatingChange, err := db.LoadLastCodeforcesRatingChangeByUID(u.DBUser.ID)
		if err != nil {
			return err
		}

		if lastRatingChange != nil {
			lastRatingChangeAt = lastRatingChange.At
		}
	}

	ratingChanges, err := fetcher.FetchCodeforcesRatingChangesAfter(handle, lastRatingChangeAt)
	if err != nil {
		return err
	}

	lastSubmitAt := time.Unix(0, 0)
	if !isNewUser {
		lastSubmission, err := db.LoadLastCodeforcesSubmissionByUID(u.DBUser.ID)
		if err != nil {
			return err
		}

		if lastSubmission != nil {
			lastSubmitAt = lastSubmission.At
		}
	}

	submissions, err := fetcher.FetchCodeforcesSubmissionsAfter(handle, lastSubmitAt)

	if err = u.fromFetcherRatingChanges(ratingChanges); err != nil {
		return err
	}

	if err = u.fromFetcherSubmissions(submissions); err != nil {
		return err
	}

	if err = u.fromFetcherUserInfo(userInfo); err != nil {
		return err
	}

	if err = u.saveUser2DB(); err != nil {
		return err
	}

	if err = u.saveFetcherRatingChanges2DB(ratingChanges); err != nil {
		return err
	}

	if err = u.saveFetcherSubmissions2DB(submissions); err != nil {
		return err
	}

	return nil
}

// process 对输出数据进行预处理
func (u *CodeforcesUser) process() (err error) {
	// Do Not show submission to outside
	// ---------------------------------------------------------------------- //
	u.DBUser.Submissions = nil
	// ---------------------------------------------------------------------- //
	// 解题数据
	// ---------------------------------------------------------------------- //
	var solvedProblems []db.CodeforcesProblem
	if solvedProblems, err = db.LoadCodeforcesSolvedProblemByUID(u.DBUser.ID); err != nil {
		return err
	}

	m := make(map[int]int)

	for _, problem := range solvedProblems {
		m[problem.Rating]++
	}

	for k, v := range m {
		u.SolvedProblems = append(u.SolvedProblems, SolvedData{
			RatingRange: k,
			Count:       v,
		})
	}

	sort.Slice(u.SolvedProblems, func(i, j int) bool {
		return u.SolvedProblems[i].RatingRange > u.SolvedProblems[j].RatingRange
	})
	// ---------------------------------------------------------------------- //

	// 从rating changes中读取maxRating和Rating, 因为有些人的这两个数据在user.info不公开，但是可以通过user.rating拿到
	// ---------------------------------------------------------------------- //
	if (u.DBUser.Rating == 0 || u.DBUser.MaxRating == 0) && len(u.DBUser.RatingChanges) > 0 {
		u.DBUser.Rating = u.DBUser.RatingChanges[len(u.DBUser.RatingChanges)-1].NewRating
		for _, change := range u.DBUser.RatingChanges {
			u.DBUser.MaxRating = max(u.DBUser.MaxRating, change.NewRating)
		}
	}
	// ---------------------------------------------------------------------- //

	u.SolvedCount, err = db.CountCodeforcesSolvedByUID(u.DBUser.ID)
	if err != nil {
		return err
	}

	u.DBUser.RatingChanges, err = db.LoadCodeforcesRatingChangesByUID(u.DBUser.ID)
	if err != nil {
		return err
	}

	return nil
}

func GetUpdatedCodeforcesUser(handle string) (user *CodeforcesUser, err error) {
	v, _ := updatingUser.LoadOrStore(handle, &sync.Mutex{})
	lock := v.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	user = &CodeforcesUser{}

	if err = user.loadFromCache(handle); err != nil && !cache.IsNil(err) {
		return nil, err
	}

	if err == nil {
		return user, nil
	}

	if err = user.cruDBUser(handle); err != nil {
		return nil, err
	}

	if err = user.process(); err != nil {
		return nil, err
	}

	if err = user.saveToCache(); err != nil {
		return nil, err
	}

	return user, nil
}
