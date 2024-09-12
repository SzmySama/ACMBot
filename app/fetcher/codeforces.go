package fetcher

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"math/rand"

	"github.com/SzmySama/ACMBot/app/model/db"
	"github.com/SzmySama/ACMBot/app/types"
	"github.com/SzmySama/ACMBot/app/utils/config"
	log "github.com/sirupsen/logrus"
)

const (
	SINGAL_FETCH_COUNT = 100
)

func fetchCodeforcesAPI[T any](apiMethod string, args map[string]any) (*T, error) {
	apiURL := "https://codeforces.com/api/"
	cfg := config.GetConfig().Codeforces

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
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res codeforcesResponse[T]
	if err := json.Unmarshal([]byte(body), &res); err != nil {
		return nil, err
	}

	if res.Status != "OK" {
		log.Warnf("Status is not OK")
		return nil, fmt.Errorf(res.Comment)
	}

	return &res.Result, nil
}

func FetchCodeforcesUsersInfo(handles []string, checkHistoricHandles bool) (user *[]types.User, err error) {
	return fetchCodeforcesAPI[[]types.User]("user.info", map[string]any{
		"handles":              strings.Join(handles, ";"),
		"checkHistoricHandles": checkHistoricHandles,
	})
}

func FetchCodeforcesUserSubmissions(handle string, from, count int) (*[]types.Submission, error) {
	return fetchCodeforcesAPI[[]types.Submission]("user.status", map[string]any{
		"handle": handle,
		"from":   from,
		"count":  count,
	})
}

func FetchCodeforcesUserRatingChanges(handle string) (*[]types.RatingChange, error) {
	return fetchCodeforcesAPI[[]types.RatingChange]("user.rating", map[string]any{
		"handle": handle,
	})
}

func UpdateCodeforcesUserSubmissions(handle string) error {
	/*
		1. 获取用户，不存在则返回
		2. 获取Submissions的更新时间
		3. fetch用户的提交记录，更新数据库相关数据
	*/
	db := db.GetDBConnection()
	var user types.User
	if result := db.Where("handle = ?", handle).First(&user); result.Error != nil {
		return fmt.Errorf("failed to find user %s in DB: %v", handle, result.Error)
	}
	// fetch

	var fetchCount = 1
	var newSubmissions []types.Submission
	var currectLastSubmissionTimeStamp time.Time

	for {
		res, err := FetchCodeforcesUserSubmissions(handle, fetchCount, SINGAL_FETCH_COUNT)
		if err != nil {
			return err
		}
		if len(*res) == 0 {
			break
		}
		currectLastSubmissionTimeStamp = (*res)[len(*res)-1].At

		if currectLastSubmissionTimeStamp.Sub(user.SubmissionUpdatedAt).Seconds() > 0 {
			// 当前获取到的数据和原始数据没有交集
			newSubmissions = append(newSubmissions, *res...)
			fetchCount += SINGAL_FETCH_COUNT
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
		if v.Status == types.SUBMISSION_STATUS_OK {
			user.Solved++
		}
	}

	user.Submissions = append(newSubmissions, user.Submissions...)
	user.SubmissionUpdatedAt = user.Submissions[0].At
	log.Infof("%v", user.Submissions)
	// 更新数据库数据
	if err := db.Save(&user).Error; err != nil {
		return err
	}
	return nil
}

func UpdateCodeforcesUserRatingChanges(handle string) error {
	db := db.GetDBConnection()
	var user types.User
	if result := db.Where("handle = ?", handle).First(&user); result.Error != nil {
		return fmt.Errorf("failed to find user %s in DB: %v", handle, result.Error)
	}

	ratingChanges, err := FetchCodeforcesUserRatingChanges(handle)
	if err != nil {
		return err
	}
	user.RatingChanges = *ratingChanges
	if length := len(*ratingChanges); length > 0 {
		user.RatingChangeUpdateAt = user.RatingChanges[length-1].At
	}
	if err := db.Save(&user).Error; err != nil {
		return err
	}
	return nil
}
