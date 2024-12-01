package fetcher

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app/model/errs"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"math/rand"

	"github.com/YourSuzumiya/ACMBot/app/utils/config"
	log "github.com/sirupsen/logrus"
)

var (
	cfg *config.CodeforcesConfigStruct
)

func init() {
	cfg = &config.GetConfig().Codeforces
}

type CodeforcesUser struct {
	Handle       string `json:"handle"`
	Avatar       string `json:"titlePhoto"`
	Rating       int    `json:"rating"`
	MaxRating    int    `json:"maxRating"`
	Solved       int
	FriendCount  int    `json:"friendOfCount"`
	Organization string `json:"organization"`
	CreatedAt    int64  `json:"registrationTimeSeconds"`
}

type CodeforcesProblem struct {
	ContestID      int      `json:"contestId"`
	ProblemSetName string   `json:"problemsetName"`
	Index          string   `json:"index"`
	Rating         int      `json:"rating"`
	Tags           []string `json:"tags"`
}

func (p *CodeforcesProblem) ID() string {
	if p.ContestID == 0 {
		return p.ProblemSetName + p.Index
	}
	return fmt.Sprintf("%d%s", p.ContestID, p.Index)
}

type CodeforcesSubmission struct {
	ID      uint              `json:"id"`
	At      int64             `json:"creationTimeSeconds"`
	Status  string            `json:"verdict"`
	Problem CodeforcesProblem `json:"problem"`
}

type CodeforcesRatingChange struct {
	At        int64 `json:"ratingUpdateTimeSeconds"`
	NewRating int   `json:"newRating"`
}

type CodeforcesRace struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	Type                string `json:"type"`
	Phase               string `json:"phase"`
	Frozen              bool   `json:"frozen"`
	DurationSeconds     int64  `json:"durationSeconds"`
	StartTimeSeconds    int64  `json:"startTimeSeconds"`
	RelativeTimeSeconds int64  `json:"relativeTimeSeconds"`
}

func zero[T any]() T {
	var t T
	return t
}

var cfLock sync.Mutex

func fetchCodeforcesAPI[T any](apiMethod string, args map[string]any) (T, error) {
	cfLock.Lock()
	defer cfLock.Unlock()
	time.Sleep(500 * time.Millisecond)
	type codeforcesResponse[T any] struct {
		/*
			codeforces响应数据的基本格式:
				Result->期望的数据
				Comment->失败时返回的提示信息
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
		return zero[T](), err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("failed to close response body: %v", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero[T](), err
	}

	var res codeforcesResponse[T]
	if err := json.Unmarshal(body, &res); err != nil {
		return zero[T](), err
	}
	if res.Status != "OK" {
		if strings.HasSuffix(res.Comment, "not found") {
			return zero[T](), errs.ErrHandleNotFound
		}
		log.Infof("Status is not OK")
		return zero[T](), fmt.Errorf(res.Comment)
	}

	return res.Result, nil
}

/*
	为什么都用slice了为什么还要返回一个指针
	因为期望类型未知，不一定是slice
*/

func FetchCodeforcesUsersInfo(handles []string, checkHistoricHandles bool) ([]CodeforcesUser, error) {
	return fetchCodeforcesAPI[[]CodeforcesUser]("user.info", map[string]any{
		"handles":              strings.Join(handles, ";"),
		"checkHistoricHandles": checkHistoricHandles,
	})
}

func FetchCodeforcesSubmissions(handle string, from, count int) ([]CodeforcesSubmission, error) {
	return fetchCodeforcesAPI[[]CodeforcesSubmission]("user.status", map[string]any{
		"handle": handle,
		"from":   from,
		"count":  count,
	})
}

func FetchCodeforcesRatingChanges(handle string) ([]CodeforcesRatingChange, error) {
	return fetchCodeforcesAPI[[]CodeforcesRatingChange]("user.rating", map[string]any{
		"handle": handle,
	})
}

func FetchCodeforcesContestList(gym bool) ([]CodeforcesRace, error) {
	return fetchCodeforcesAPI[[]CodeforcesRace]("contest.list", map[string]any{
		"gym": gym,
	})
}

func FetchCodeforcesUserInfo(handle string, checkHistoricHandles bool) (*CodeforcesUser, error) {
	users, err := FetchCodeforcesUsersInfo([]string{handle}, checkHistoricHandles)
	if err != nil {
		return nil, err
	}
	if len(users) != 1 {
		return nil, errs.ErrHandleNotFound
	}
	return &users[0], nil
}

func FetchCodeforcesRatingChangesAfter(handle string, after time.Time) ([]CodeforcesRatingChange, error) {
	allChanges, err := FetchCodeforcesRatingChanges(handle)
	if err != nil {
		return nil, err
	}
	if after.Unix() == 0 {
		return allChanges, nil
	}
	for i, change := range allChanges {
		if time.Unix(change.At, 0).After(after) {
			return allChanges[i:], nil
		}
	}
	return []CodeforcesRatingChange{}, nil
}

func FetchCodeforcesSubmissionsAfter(handle string, after time.Time) ([]CodeforcesSubmission, error) {
	perFetch := 500
	if after.Unix() == 0 {
		perFetch = 10000
	}
	allSubmissions := make([]CodeforcesSubmission, 0, perFetch)
	count := 1
	for {
		correct, err := FetchCodeforcesSubmissions(handle, count, perFetch)
		if err != nil {
			return nil, err
		}
		if len(correct) == 0 {
			break
		}
		correctStart := time.Unix(correct[0].At, 0)            // 最晚的submission
		correctEnd := time.Unix(correct[len(correct)-1].At, 0) // 最早的submission
		// 所有submission都早于期望时间
		if correctStart.Before(after) {
			break
			// 有部分submission早于期望时间
		} else if correctEnd.Before(after) {
			for _, submission := range correct {
				// 早于或等于的都不要
				if !time.Unix(submission.At, 0).After(after) {
					break
				}
				allSubmissions = append(allSubmissions, submission)
			}
			break
			// 全部submission都在期望时间之后
		} else {
			allSubmissions = append(allSubmissions, correct...)
		}
		if len(correct) < perFetch {
			break
		}
		count += perFetch
	}
	return allSubmissions, nil
}
