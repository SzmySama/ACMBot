package manager

import (
	"sort"
	"sync"
	"time"

	"encoding/json"

	"github.com/YourSuzumiya/ACMBot/app/model/cache"
	"github.com/YourSuzumiya/ACMBot/app/model/db"
	"github.com/YourSuzumiya/ACMBot/app/model/fetcher"
	"github.com/YourSuzumiya/ACMBot/app/model/render"
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
	RatingRange uint
	Count       uint
}

/*
CodeforcesUser
必须preload所有RatingChanges和最后一条Submission
才能执行相关函数
*/
type CodeforcesUser struct {
	DBUser         db.CodeforcesUser
	SolvedProblems []SolvedData
}

func (u *CodeforcesUser) LoadFromDB(handle string) error {
	if user, err := db.LoadCodeforcesUserByHandle(handle); err != nil {
		return err
	} else {
		u.DBUser = *user
		return nil
	}
}

func (u *CodeforcesUser) FromFetcherUserInfo(user fetcher.CodeforcesUser) error {
	u.DBUser.Handle = user.Handle
	u.DBUser.Avatar = user.Avatar
	u.DBUser.Rating = user.Rating
	u.DBUser.FriendCount = user.FriendCount
	u.DBUser.MaxRating = user.MaxRating
	u.DBUser.CreatedAt = time.Unix(user.CreatedAt, 0)
	return nil
}

func (u *CodeforcesUser) FromFetcherRatingChanges(changes []fetcher.CodeforcesRatingChange) error {
	lastRatingChangeInDB := time.Unix(0, 0)
	if len(u.DBUser.RatingChanges) > 0 {
		lastRatingChangeInDB = u.DBUser.RatingChanges[len(u.DBUser.RatingChanges)-1].At
	}

	firstNewRatingChangeIdx := -1

	for idx, change := range changes {
		if time.Unix(change.At, 0).After(lastRatingChangeInDB) {
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
			At:               time.Unix(change.At, 0),
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
		if !time.Unix(submissions[i].At, 0).After(lastSubmissionInDB) {
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

func (u *CodeforcesUser) SaveToDB() error {
	//每次最多插入5000条submission
	const mx = 5000
	var err error
	var submissions []db.CodeforcesSubmission
	if len(u.DBUser.Submissions) > mx {
		submissions = u.DBUser.Submissions
		u.DBUser.Submissions = []db.CodeforcesSubmission{}
	}
	if err = db.SaveCodeforcesUser(&u.DBUser); err != nil {
		return err
	}

	if submissions != nil {
		for i := range submissions {
			submissions[i].CodeforcesUserID = u.DBUser.ID
		}

		for i := 0; i < len(submissions); i += mx {
			if err = db.SaveCodeforcesSubmissions(submissions[i:min(len(submissions), i+mx)]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *CodeforcesUser) ToRenderProfileV1() *render.CodeforcesUser {
	return &render.CodeforcesUser{
		Handle:   u.DBUser.Handle,
		Avatar:   u.DBUser.Avatar,
		Rating:   u.DBUser.Rating,
		Solved:   u.DBUser.Solved,
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
		result[k].Percent = float32(v.Count) / float32(u.DBUser.Solved) * 100
	}

	return &render.CodeforcesUserProfile{
		Avatar:     u.DBUser.Avatar,
		Handle:     u.DBUser.Handle,
		MaxRating:  u.DBUser.MaxRating,
		FriendOf:   u.DBUser.FriendCount,
		Rating:     u.DBUser.Rating,
		Solved:     u.DBUser.Solved,
		Level:      render.ConvertRatingToLevel(u.DBUser.Rating),
		SolvedData: result,
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
	if err != nil && !cache.IsNil(err) {
		return nil, err
	}

	result := CodeforcesUser{}
	if cacheUser != "" && json.Unmarshal([]byte(cacheUser), &result) == nil {
		return &result, nil
	}

	isNewUser := false
	if err = result.LoadFromDB(handle); err != nil {
		if !db.IsNotFound(err) {
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
		flag = time.Unix((*correctSubmissions)[0].At, 0).Before(lastSubmissionInDB)
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
		result.DBUser.Solved = len(solved)
	} else {
		solved, err := db.CountCodeforcesSolvedByUID(result.DBUser.ID)
		if err != nil {
			return nil, err
		}
		result.DBUser.Solved = solved
	}

	if err = result.SaveToDB(); err != nil {
		return nil, err
	}

	// Do Not show submission to outside
	result.DBUser.Submissions = nil

	// Process data

	solvedProblems, err := db.LoadCodeforcesSolvedProblemByUID(result.DBUser.ID)
	if err != nil {
		return nil, err
	}

	m := make(map[uint]uint)

	for _, problem := range solvedProblems {
		rg := uint(problem.Rating)
		if _, ok := m[rg]; ok {
			m[rg]++
		} else {
			m[rg] = 1
		}
	}

	for k, v := range m {
		result.SolvedProblems = append(result.SolvedProblems, SolvedData{
			RatingRange: k,
			Count:       v,
		})
	}

	sort.Slice(result.SolvedProblems, func(i, j int) bool {
		return result.SolvedProblems[i].RatingRange > result.SolvedProblems[j].RatingRange
	})

	// save to redis

	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	if err = cache.SetCodeforcesUser(handle, data, 4*time.Hour); err != nil {
		return nil, err
	}

	return &result, nil
}
