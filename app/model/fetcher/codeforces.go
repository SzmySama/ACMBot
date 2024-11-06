package fetcher

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	zero "github.com/wdvxdr1123/ZeroBot"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"math/rand"

	"github.com/YourSuzumiya/ACMBot/app/model/db"
	"github.com/YourSuzumiya/ACMBot/app/utils/config"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	NewUserSignalFetchCount = 10000 // 单次查询Codeforces用户的Submission的数量
	SignalFetchCount        = 500
)

var (
	dbc *gorm.DB // db connection
	cfg *config.CodeforcesConfigStruct

	updatingUser sync.Map
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
	DurationSeconds     int64  `json:"durationSeconds"`
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

var (
	cfLock sync.Mutex
)

func fetchCodeforcesAPI[T any](apiMethod string, args map[string]any) (*T, error) {
	cfLock.Lock()
	defer cfLock.Unlock()
	time.Sleep(1 * time.Second)
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

func UpdateDBCodeforcesUser(handle string, ctx *zero.Ctx) error {
	mu := &sync.Mutex{}
	loadedMu, _ := updatingUser.LoadOrStore(handle, mu)
	mu = loadedMu.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()
	newUser := false
	var dbUser db.CodeforcesUser
	if result := dbc.
		Preload("Submissions", func(db *gorm.DB) *gorm.DB {
			return db.Order("at DESC").Limit(1)
		}).
		Preload("RatingChanges").Where("handle = ?", handle).First(&dbUser); result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}
		if err := CreateDBCodeforcesUser(handle); err != nil {
			return fmt.Errorf("failed to update DB codeforces user: %v", err)
		}
		newUser = true
	}

	if time.Since(dbUser.UpdatedAt).Hours() <= 4 {
		return nil
	}

	if newUser {
		if err := dbc.Where("handle = ?", handle).First(&dbUser).Error; err != nil {
			return fmt.Errorf("failed to update DB codeforces user: %v", err)
		}
	}

	log.Info("updating DB codeforces user")
	if ctx != nil {
		ctx.Send("正在更新用户数据，请稍后...")
	}
	var err error
	if !newUser {
		if err = UpdateDBCodeforcesUserInfo(&dbUser); err != nil {
			return fmt.Errorf("failed to update DB codeforces user: %v", err)
		}
	}
	if err = UpdateDBCodeforcesSubmissions(&dbUser); err != nil {
		return fmt.Errorf("failed to update DB codeforces submissions: %v", err)
	}
	if err = UpdateDBCodeforcesRatingChanges(&dbUser); err != nil {
		return fmt.Errorf("failed to update DB codeforces rating changes: %v", err)
	}

	return dbc.Save(&dbUser).Error
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

// UpdateDBCodeforcesSubmissions 需要preload最后一条submission !important
func UpdateDBCodeforcesSubmissions(user *db.CodeforcesUser) error {
	submissionNumPerFetch := SignalFetchCount
	if len(user.Submissions) == 0 {
		submissionNumPerFetch = NewUserSignalFetchCount
	}

	fetchCount := 1
	var newSubmissions []codeforcesSubmission
	lastSubmitTimeInDB := time.Unix(0, 0)
	if len(user.Submissions) > 0 {
		lastSubmitTimeInDB = user.Submissions[len(user.Submissions)-1].CreatedAt
	}

	for {
		res, err := FetchCodeforcesUserSubmissions(user.Handle, fetchCount, submissionNumPerFetch)
		if err != nil {
			return err
		}

		if len(*res) == 0 {
			break
		}

		fetchCount += submissionNumPerFetch
		lastSubmitTimeInRes := (*res)[len(*res)-1].At

		if lastSubmitTimeInRes.After(lastSubmitTimeInDB) {
			newSubmissions = append(newSubmissions, *res...)
			continue
		}

		for _, v := range *res {
			if !v.At.After(lastSubmitTimeInDB) {
				break
			}
			newSubmissions = append(newSubmissions, v)
		}
		break
	}

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
			CodeforcesUserID:    user.ID,
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

func UpdateDBCodeforcesRatingChanges(user *db.CodeforcesUser) error {
	fetchRatingChanges, err := FetchCodeforcesUserRatingChanges(user.Handle)
	if err != nil {
		return err
	}
	var lastDBRatingChange db.CodeforcesRatingChange
	if len(user.RatingChanges) > 0 {
		lastDBRatingChange = user.RatingChanges[len(user.RatingChanges)-1]
	} else {
		lastDBRatingChange = db.CodeforcesRatingChange{
			At: time.Unix(0, 0),
		}
	}

	// todo: 使用二分查找
	firstNewRatingChangeIndex := -1
	for k, v := range *fetchRatingChanges {
		if v.At.After(lastDBRatingChange.At) {
			firstNewRatingChangeIndex = k
			break
		}
	}

	if firstNewRatingChangeIndex == -1 {
		return nil
	}

	for _, v := range (*fetchRatingChanges)[firstNewRatingChangeIndex:] {
		user.RatingChanges = append(user.RatingChanges, db.CodeforcesRatingChange{
			CodeforcesUserID: user.ID,
			At:               v.At,
			NewRating:        v.NewRating,
		})
	}

	return nil
}

func UpdateDBCodeforcesUserInfo(user *db.CodeforcesUser) error {
	fetchUsers, err := FetchCodeforcesUsersInfo([]string{user.Handle}, false)
	if err != nil {
		return err
	}
	if len(*fetchUsers) != 1 {
		return fmt.Errorf("got %d user(s) instead of 1", len(*fetchUsers))
	}
	fetchUser := (*fetchUsers)[0]

	user.Handle = fetchUser.Handle
	user.Avatar = fetchUser.Avatar
	user.FriendCount = fetchUser.FriendCount
	user.Rating = fetchUser.Rating
	user.CreatedAt = fetchUser.CreatedAt

	return nil
}

/* ------------------------------------------- */
