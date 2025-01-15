package manager

import (
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/YourSuzumiya/ACMBot/app/fetcher"
	"github.com/YourSuzumiya/ACMBot/app/render"

	"github.com/YourSuzumiya/ACMBot/app/model/cache"
	"github.com/YourSuzumiya/ACMBot/app/model/db"
)

var updatingAtcoderUserUser sync.Map

type AtcoderSolvedData struct {
	RatingRange uint
	Count       uint
}

type AtcoderUser struct {
	DBUser         db.AtcoderUser
	SolvedProblems []AtcoderSolvedData
	SolvedCount    uint
}

func (u *AtcoderUser) ToRenderProfile() *render.AtcoderUserProfile {
	result := make([]render.AtcoderSolvedData, 6)
	result[0].Range = "400"
	result[1].Range = "800"
	result[2].Range = "1200"
	result[3].Range = "1600"
	result[4].Range = "1600+"
	result[5].Range = "unclassified"

	for _, problem := range u.SolvedProblems {
		switch {
		case problem.RatingRange == 0:
			result[5].Count += problem.Count
		case problem.RatingRange == 400:
			result[0].Count += problem.Count
		case problem.RatingRange < 800:
			result[1].Count += problem.Count
		case problem.RatingRange < 1200:
			result[2].Count += problem.Count
		case problem.RatingRange < 1600:
			result[3].Count += problem.Count
		default:
			result[4].Count += problem.Count
		}
	}

	for k, v := range result {
		result[k].Percent = float64(v.Count) / float64(u.SolvedCount) * 100
	}
	return &render.AtcoderUserProfile{
		Avatar:           u.DBUser.Avatar,
		Handle:           u.DBUser.Handle,
		MaxRating:        u.DBUser.MaxRating,
		Rating:           u.DBUser.Rating,
		Level:            u.DBUser.Level,
		PromotionMessage: u.DBUser.PromotionMessage,
		Solved:           u.SolvedCount,
        Time:        time.Now().Format("2006-01-02 15:04:05"),
		SolvedData:       result,
	}
}

func (u *AtcoderUser) fromFetcherUserInfo(user *fetcher.AtcoderUser) error {
	u.DBUser.Handle = user.Handle
	u.DBUser.Avatar = user.Avatar
	u.DBUser.Rating = user.Rating
	u.DBUser.MaxRating = user.HighestRating
	u.DBUser.Level = user.Dan
	u.DBUser.PromotionMessage = user.PromotionMessage[1 : len(user.PromotionMessage)-1]
	u.DBUser.CreatedAt = time.Now()
	return nil
}

func (u *AtcoderUser) fromFetcherSubmissions(submissions []fetcher.AtcoderSubmission) error {
	problemID2Submission := make(map[string][]db.AtcoderSubmission)
	problemID2Problem := make(map[string]db.AtcoderProblem)
	newSubmissions := make([]db.AtcoderSubmission, 0, len(submissions))

	for _, submission := range submissions {
		id := submission.ProblemId
		dbSubmission := db.AtcoderSubmission{
			AtcoderUserID:    u.DBUser.ID,
			AtcoderProblemID: id,
			SubmissionTime:   time.Unix(submission.SubmissionTime, 0),
			Status:           submission.Status,
		}

		newSubmissions = append(newSubmissions, dbSubmission)
		problemID2Submission[id] = append(problemID2Submission[id], dbSubmission)
		problemID2Problem[id] = db.AtcoderProblem{
			ID:     id,
			Rating: uint(submission.Point),
		}
	}

	s := make([]db.AtcoderProblem, 0, len(problemID2Problem))
	for id := range problemID2Submission {
		s = append(s, db.AtcoderProblem{
			ID:     id,
			Rating: problemID2Problem[id].Rating,
		})
	}

	if len(s) > 0 {
		if err := db.SaveAtcoderProblems(s); err != nil {
			return err
		}
	}

	u.DBUser.Submissions = append(u.DBUser.Submissions, newSubmissions...)
	return nil
}

func (u *AtcoderUser) loadFromDB(handle string) error {
	if user, err := db.LoadAtcoderUserByHandle(handle); err != nil {
		return err
	} else {
		u.DBUser = *user
		return nil
	}
}

func (u *AtcoderUser) saveUser2DB() error {
	user := u.DBUser
	user.Submissions = nil
	err := db.SaveAtcoderUser(&user)
	if err != nil {
		return err
	}
	u.DBUser.ID = user.ID
	return nil
}

func (u *AtcoderUser) saveFetcherSubmissions2DB(submissions []fetcher.AtcoderSubmission) error {
	dbSubmissions := make([]db.AtcoderSubmission, 0, len(submissions))
	for _, submission := range submissions {
		dbSubmissions = append(dbSubmissions, db.AtcoderSubmission{
			AtcoderUserID:    u.DBUser.ID,
			AtcoderProblemID: submission.ProblemId,
			SubmissionTime:   time.Unix(submission.SubmissionTime, 0),
			Status:           submission.Status,
		})
	}
	return db.SaveAtcoderSubmissions(dbSubmissions)
}

func (u *AtcoderUser) loadFromCache(handle string) (err error) {
	var data string
	if data, err = cache.GetAtcoderUser(handle); err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), &u)
}

func (u *AtcoderUser) saveToCache() (err error) {
	var data []byte
	if data, err = json.Marshal(u); err != nil {
		return err
	}
	if err = cache.SetAtcoderUser(u.DBUser.Handle, data, 4*time.Hour); err != nil {
		return err
	}
	return nil
}

// cruDB create/read & update user data in DB
func (u *AtcoderUser) cruDBUser(handle string) (err error) {
	isNewUser := false
	if err = u.loadFromDB(handle); err != nil {
		if !db.IsNotFound(err) {
			return err
		}
		isNewUser = true
	}

	userInfo, err := fetcher.FetchAtcoderUser(handle)
	if err != nil {
		return err
	}

	lastSubmissionTime := time.Unix(0, 0)
	if !isNewUser {
		lastSubmission, err := db.LoadLastAtcoderSubmissionByUID(u.DBUser.ID)
		if err != nil {
			return err
		}

		if lastSubmission != nil {
			lastSubmissionTime = lastSubmission.SubmissionTime
		}
	}

	submissions, err := fetcher.FetchAtcoderSubmissionListFrom(handle, lastSubmissionTime.Unix())
	if err != nil {
		return err
	}

	if err = u.fromFetcherSubmissions(*submissions); err != nil {
		return err
	}

	if err = u.fromFetcherUserInfo(userInfo); err != nil {
		return err
	}

	if err = u.saveUser2DB(); err != nil {
		return err
	}

	if err = u.saveFetcherSubmissions2DB(*submissions); err != nil {
		return err
	}

	return nil
}

// process 对输出数据进行预处理
func (u *AtcoderUser) process() (err error) {
	// Do Not show submission to outside
	// ---------------------------------------------------------------------- //
	u.DBUser.Submissions = nil
	// ---------------------------------------------------------------------- //
	// 解题数据
	// ---------------------------------------------------------------------- //
	var solvedProblems []db.AtcoderProblem
	if solvedProblems, err = db.LoadAtcoderSolvedProblemByUID(u.DBUser.ID); err != nil {
		return err
	}

	m := make(map[uint]uint)

	for _, problem := range solvedProblems {
		m[problem.Rating]++
	}

	for k, v := range m {
		u.SolvedProblems = append(u.SolvedProblems, AtcoderSolvedData{
			RatingRange: k,
			Count:       v,
		})
	}

	sort.Slice(u.SolvedProblems, func(i, j int) bool {
		return u.SolvedProblems[i].RatingRange > u.SolvedProblems[j].RatingRange
	})

	u.SolvedCount, err = db.CountAtcoderSolvedByUID(u.DBUser.ID)
	return err
}

func GetUpdatedAtcoderUser(handle string) (user *AtcoderUser, err error) {
	v, _ := updatingAtcoderUserUser.LoadOrStore(handle, &sync.Mutex{})
	lock := v.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	user = &AtcoderUser{}
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
