package fetcher

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"math/rand"

	"github.com/YourSuzumiya/ACMBot/app/model/db"
	"github.com/YourSuzumiya/ACMBot/app/types"
	"github.com/YourSuzumiya/ACMBot/app/utils/config"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	SignalFetchCount = 500 // 单次查询Codeforces用户的Submission的数量
)

var (
	dbc *gorm.DB // db connection
	cfg *config.CodeforcesConfigStruct
)

func init() {
	dbc = db.GetDBConnection()
	cfg = &config.GetConfig().Codeforces
}

type codeforcesUser struct {
	Handle      string `json:"handle"`
	Avatar      string `json:"avatar"`
	Rating      uint   `json:"rating"`
	Solved      uint
	FriendCount uint      `json:"friendOfCount"`
	CreatedAt   time.Time `json:"-"`
}

func (u *codeforcesUser) UnmarshalJSON(data []byte) error {
	type alias codeforcesUser
	aux := &struct {
		T int64 `json:"registrationTimeSeconds"`
		*alias
	}{
		alias: (*alias)(u),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	u.CreatedAt = time.Unix(aux.T, 0)
	return nil
}

type codeforcesProblem struct {
	ContestID      int      `json:"contestId"`
	ProblemSetName string   `json:"problemsetName"`
	Index          string   `json:"index"`
	Rating         int      `json:"rating"`
	Tags           []string `json:"tags"`
}

func (p *codeforcesProblem) ID() string {
	if p.ContestID == 0 {
		return p.ProblemSetName + p.Index
	}
	return fmt.Sprintf("%d%s", p.ContestID, p.Index)
}

type codeforcesSubmission struct {
	ID      uint              `json:"id"`
	At      time.Time         `json:"-"`
	Status  string            `json:"verdict"`
	Problem codeforcesProblem `json:"problem"`
}

func (s *codeforcesSubmission) UnmarshalJSON(data []byte) error {
	type alias codeforcesSubmission
	aux := &struct {
		T int64 `json:"creationTimeSeconds"`
		*alias
	}{
		alias: (*alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	s.At = time.Unix(aux.T, 0)
	return nil
}

type codeforcesRatingChange struct {
	At        time.Time `json:"-"`
	NewRating int       `json:"newRating"`
}

type CodeforcesRace struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	Type                string `json:"type"`
	Phase               string `json:"phase"`
	Frozen              bool   `json:"frozen"`
	DurationSeconds     int    `json:"durationSeconds"`
	StartTimeSeconds    int64  `json:"startTimeSeconds"`
	RelativeTimeSeconds int    `json:"relativeTimeSeconds"`
}

func (r *codeforcesRatingChange) UnmarshalJSON(data []byte) error {
	type alias codeforcesRatingChange
	aux := &struct {
		T int64 `json:"ratingUpdateTimeSeconds"`
		*alias
	}{
		alias: (*alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.At = time.Unix(aux.T, 0)
	return nil
}

func fetchCodeforcesAPI[T any](apiMethod string, args map[string]any) (*T, error) {

	type codeforcesResponse[T any] struct {
		/*
			codeforces响应数据的基本格式
			Result是期望的数据
			Comment是失败时返回的提示信息
		*/
		Status  string `json:"status"`
		Result  T      `json:"result"`
		Comment string `json:"comment"`
	}

	apiURL := "https://codeforces.com/api/"

	args["apiKey"] = cfg.Key
	args["time"] = strconv.Itoa(int(time.Now().Unix()))

	var sortedArgs []string
	for k, v := range args {
		sortedArgs = append(sortedArgs, fmt.Sprintf("%v=%v", k, v))
	}
	sort.Strings(sortedArgs)

	randStr := strconv.Itoa(rand.Intn(900000) + 100000)
	hashSource := randStr + "/" + apiMethod + "?" + strings.Join(sortedArgs, "&") + "#" + cfg.Secret

	h := sha512.New()
	h.Write([]byte(hashSource))
	hashSig := hex.EncodeToString(h.Sum(nil))

	apiFullURL := apiURL + apiMethod + "?"
	for _, arg := range sortedArgs {
		apiFullURL += arg + "&"
	}
	apiFullURL += "apiSig=" + randStr + hashSig

	log.Infof(apiFullURL)

	resp, err := http.Get(apiFullURL)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res codeforcesResponse[T]
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	if res.Status != "OK" {
		log.Warnf("Status is not OK")
		return nil, fmt.Errorf(res.Comment)
	}

	return &res.Result, nil
}

/*
	为什么都用slice了为什么还要返回一个指针呢
	因为期望类型未知，不一定是slice
*/

func FetchCodeforcesUsersInfo(handles []string, checkHistoricHandles bool) (*[]codeforcesUser, error) {
	return fetchCodeforcesAPI[[]codeforcesUser]("user.info", map[string]any{
		"handles":              strings.Join(handles, ";"),
		"checkHistoricHandles": checkHistoricHandles,
	})
}

func FetchCodeforcesUserSubmissions(handle string, from, count int) (*[]codeforcesSubmission, error) {
	return fetchCodeforcesAPI[[]codeforcesSubmission]("user.status", map[string]any{
		"handle": handle,
		"from":   from,
		"count":  count,
	})
}

func FetchCodeforcesUserRatingChanges(handle string) (*[]codeforcesRatingChange, error) {
	return fetchCodeforcesAPI[[]codeforcesRatingChange]("user.rating", map[string]any{
		"handle": handle,
	})
}

func FetchCodeforcesContestList(gym bool) (*[]CodeforcesRace, error) {
	return fetchCodeforcesAPI[[]CodeforcesRace]("contest.list", map[string]any{
		"gym": gym,
	})
}

// CreateDBCodeforcesUser 当且仅当已确认无此用户时使用
func CreateDBCodeforcesUser(handle string) error {
	fetchUsers, err := FetchCodeforcesUsersInfo([]string{handle}, false)
	if err != nil {
		return err
	}

	if len(*fetchUsers) != 1 {
		return fmt.Errorf("got %d user(s) instead of 1", len(*fetchUsers))
	}
	fetchUser := (*fetchUsers)[0]
	var dbUser db.CodeforcesUser

	dbUser.Handle = fetchUser.Handle
	dbUser.Avatar = fetchUser.Avatar
	dbUser.FriendCount = fetchUser.FriendCount
	dbUser.Rating = fetchUser.Rating
	dbUser.CreatedAt = fetchUser.CreatedAt

	return dbc.Save(&dbUser).Error
}

func UpdateDBCodeforcesUserInfo(handle string) error {
	fetchUsers, err := FetchCodeforcesUsersInfo([]string{handle}, false)
	if err != nil {
		return err
	}
	if len(*fetchUsers) != 1 {
		return fmt.Errorf("got %d user(s) instead of 1", len(*fetchUsers))
	}
	fetchUser := (*fetchUsers)[0]
	var dbUser db.CodeforcesUser
	if result := dbc.Where("handle = ?", handle).First(&dbUser); result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	dbUser.Handle = fetchUser.Handle
	dbUser.Avatar = fetchUser.Avatar
	dbUser.FriendCount = fetchUser.FriendCount
	dbUser.Rating = fetchUser.Rating
	dbUser.CreatedAt = fetchUser.CreatedAt

	return dbc.Save(&dbUser).Error
}

func UpdateDBCodeforcesRatingChanges(handle string) error {
	fetchRatingChanges, err := FetchCodeforcesUserRatingChanges(handle)
	if err != nil {
		return err
	}

	var dbUser db.CodeforcesUser
	if result := dbc.Preload("RatingChanges", func(db *gorm.DB) *gorm.DB {
		return db.Order("at DESC").Limit(1)
	}).Where("handle = ?", handle).First(&dbUser); result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}

		if err = CreateDBCodeforcesUser(handle); err != nil {
			return fmt.Errorf("failed to update DB codeforces user: %v", err)
		}

		if err = dbc.Preload("RatingChanges").Where("handle = ?", handle).First(&dbUser).Error; err != nil {
			return fmt.Errorf("failed to update DB codeforces user: %v", err)
		}
	}

	var lastDBRatingChange db.CodeforcesRatingChange
	if len(dbUser.RatingChanges) > 0 {
		lastDBRatingChange = dbUser.RatingChanges[len(dbUser.RatingChanges)-1]
	} else {
		lastDBRatingChange = db.CodeforcesRatingChange{
			At: time.Unix(0, 0),
		}
	}

	// todo: 使用二分查找
	firstNewRatingChangeIndex := 0
	for k, v := range *fetchRatingChanges {
		if v.At.After(lastDBRatingChange.At) {
			firstNewRatingChangeIndex = k
			break
		}
	}

	for _, v := range (*fetchRatingChanges)[firstNewRatingChangeIndex:] {
		dbUser.RatingChanges = append(dbUser.RatingChanges, db.CodeforcesRatingChange{
			CodeforcesUserID: dbUser.ID,
			At:               v.At,
			NewRating:        v.NewRating,
		})
	}
	return dbc.Save(&dbUser).Error
}

func UpdateDBCodeforcesSubmissions(handle string) error {
	var dbUser db.CodeforcesUser
	if result := dbc.Preload("Submissions", func(db *gorm.DB) *gorm.DB {
		return db.Order("at DESC").Limit(1)
	}).Where("handle = ?", handle).First(&dbUser); result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}
		if err := CreateDBCodeforcesUser(handle); err != nil {
			return fmt.Errorf("failed to update DB codeforces user: %v", err)
		}
		if result := dbc.Preload("Submissions", func(db *gorm.DB) *gorm.DB {
			return db.Order("at DESC").Limit(1)
		}).Where("handle = ?", handle).First(&dbUser); result.Error != nil {
			return result.Error
		}
	}

	fetchCount := 1
	var newSubmissions []codeforcesSubmission
	lastSubmitTimeInDB := time.Unix(0, 0)
	if len(dbUser.Submissions) > 0 {
		lastSubmitTimeInDB = dbUser.Submissions[len(dbUser.Submissions)-1].CreatedAt
	}

	for {
		res, err := FetchCodeforcesUserSubmissions(handle, fetchCount, SignalFetchCount)
		if err != nil {
			return err
		}

		if len(*res) == 0 {
			break
		}

		fetchCount += SignalFetchCount
		lastSubmitTimeInRes := (*res)[len(*res)-1].At

		if lastSubmitTimeInRes.After(lastSubmitTimeInDB) {
			newSubmissions = append(newSubmissions, *res...)
			continue
		}

		for _, v := range *res {
			if !v.At.After(lastSubmitTimeInRes) {
				break
			}
			newSubmissions = append(newSubmissions, v)
		}
		break
	}

	// 此处已拿到所有Submission，需要将每个Submission对应到具体的Problem

	dbProblems := make(map[string]*db.CodeforcesProblem)
	for _, v := range newSubmissions {
		problemID := v.Problem.ID()
		p, ok := dbProblems[problemID]
		if !ok {
			p = &db.CodeforcesProblem{}
			result := dbc.Where("id = ?", problemID).First(p)
			if result.Error != nil {
				if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
					return fmt.Errorf("DBErr: failed to read DB codeforces problem: %v", result.Error)
				}

				p.ID = problemID
				p.Rating = v.Problem.Rating

				dbc.Create(p)
			}
			dbProblems[problemID] = p
		}
		p.Submissions = append(p.Submissions, db.CodeforcesSubmission{
			Model: gorm.Model{
				ID: v.ID,
			},
			CodeforcesUserID:    dbUser.ID,
			CodeforcesProblemID: problemID,
			At:                  v.At,
			Status:              v.Status,
		})
	}

	for _, v := range dbProblems {

		if result := dbc.Save(v.Submissions); result.Error != nil {
			return fmt.Errorf("failed to update DB codeforces submission: %v", result.Error)
		}
	}

	return nil
}

/* ------------------------------------------- */

func FetchCodeforcesUsersInfo_(handles []string, checkHistoricHandles bool) (user *[]types.User, err error) {
	return fetchCodeforcesAPI[[]types.User]("user.info", map[string]any{
		"handles":              strings.Join(handles, ";"),
		"checkHistoricHandles": checkHistoricHandles,
	})
}

func FetchCodeforcesUserSubmissions_(handle string, from, count int) (*[]types.Submission, error) {
	return fetchCodeforcesAPI[[]types.Submission]("user.status", map[string]any{
		"handle": handle,
		"from":   from,
		"count":  count,
	})
}

func FetchCodeforcesUserRatingChanges_(handle string) (*[]types.RatingChange, error) {
	return fetchCodeforcesAPI[[]types.RatingChange]("user.rating", map[string]any{
		"handle": handle,
	})
}

func FetchCodeforcesContestList_(gym bool) (*[]CodeforcesRace, error) {
	return fetchCodeforcesAPI[[]CodeforcesRace]("contest.list", map[string]any{
		"gym": gym,
	})
}

func UpdateCodeforcesUserSubmissionsAndRating_(handle string) error {
	/*
		1. 获取用户，不存在则返回
		2. 获取Submissions的更新时间
		3. fetch用户的提交记录，更新数据库相关数据
	*/
	dbConnection := db.GetDBConnection()
	var user types.User
	if result := dbConnection.Where("handle = ?", handle).First(&user); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if err := UpdateCodeforcesUserInfo_(handle); err != nil {
				return err
			}
			if err := dbConnection.Where("handle = ?", handle).First(&user).Error; err != nil {
				return fmt.Errorf("panic err while fetch user: Unexpected brach: %v", err)
			}
		} else {
			return fmt.Errorf("failed to find user %s in DB: %v", handle, result.Error)
		}
	}
	// fetch

	if time.Since(user.SubmissionUpdatedAt).Hours() <= 24 {
		return nil
	}

	var fetchCount = 1
	var newSubmissions []types.Submission
	var correctLastSubmissionTimeStamp time.Time

	for {
		res, err := FetchCodeforcesUserSubmissions_(handle, fetchCount, SignalFetchCount)
		if err != nil {
			return err
		}
		if len(*res) == 0 {
			break
		}
		correctLastSubmissionTimeStamp = (*res)[len(*res)-1].At

		if correctLastSubmissionTimeStamp.Sub(user.SubmissionUpdatedAt).Seconds() > 0 {
			// 当前获取到的数据和原始数据没有交集
			newSubmissions = append(newSubmissions, *res...)
			fetchCount += SignalFetchCount
			continue
		} else {
			for _, v := range *res {
				if v.At.Sub(user.SubmissionUpdatedAt).Seconds() <= 0 {
					break
				}
				newSubmissions = append(newSubmissions, v)
			}
			break
		}
	}

	for _, v := range newSubmissions {
		if v.Status == types.SubmissionStatusOk {
			user.Solved++
		}
	}

	user.Submissions = append(newSubmissions, user.Submissions...)
	user.SubmissionUpdatedAt = time.Now()

	// 更新rating数据
	if u, err := FetchCodeforcesUsersInfo_([]string{handle}, false); err != nil {
		return fmt.Errorf("failed to update cf user: %v", err)
	} else {
		user.Rating = (*u)[0].Rating
	}

	// 更新数据库数据
	if err := dbConnection.Save(&user).Error; err != nil {
		return err
	}
	return nil
}

func UpdateCodeforcesUserRatingChanges_(handle string) error {
	dbc := db.GetDBConnection()
	var user types.User
	if result := dbc.Where("handle = ?", handle).First(&user); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if err := UpdateCodeforcesUserInfo_(handle); err != nil {
				return err
			}

			if err := dbc.Where("handle = ?", handle).First(&user).Error; err != nil {
				return fmt.Errorf("panic err while fetch user: Unexpected brach: %v", err)
			}

		} else {
			return fmt.Errorf("failed to find user %s in DB: %v", handle, result.Error)
		}
	}

	if time.Since(user.RatingChangeUpdateAt) <= 30*time.Minute {
		return nil
	}

	ratingChanges, err := FetchCodeforcesUserRatingChanges_(handle)
	if err != nil {
		return err
	}
	user.RatingChanges = *ratingChanges
	if length := len(*ratingChanges); length > 0 {
		user.RatingChangeUpdateAt = time.Now()
	}
	if err := dbc.Save(&user).Error; err != nil {
		return err
	}
	return nil
}

func UpdateCodeforcesUserInfo_(handle string) error {

	user, err := FetchCodeforcesUsersInfo_([]string{handle}, false)
	if err != nil {
		return fmt.Errorf("failed to update cf user: %v", err)
	}
	return db.GetDBConnection().Save(&((*user)[0])).Error
}
