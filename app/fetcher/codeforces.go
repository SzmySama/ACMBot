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

	"github.com/SzmySama/ACMBot/app/model/db"
	"github.com/SzmySama/ACMBot/app/types"
	"github.com/SzmySama/ACMBot/app/utils/config"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	SignalFetchCount = 500
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

func FetchCodeforcesContestList(gym bool) (*[]CodeforcesRace, error) {
	return fetchCodeforcesAPI[[]CodeforcesRace]("contest.list", map[string]any{
		"gym": gym,
	})
}

func UpdateCodeforcesUserSubmissionsAndRating(handle string) error {
	/*
		1. 获取用户，不存在则返回
		2. 获取Submissions的更新时间
		3. fetch用户的提交记录，更新数据库相关数据
	*/
	dbConnection := db.GetDBConnection()
	var user types.User
	if result := dbConnection.Where("handle = ?", handle).First(&user); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if err := UpdateCodeforcesUserInfo(handle); err != nil {
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
		res, err := FetchCodeforcesUserSubmissions(handle, fetchCount, SignalFetchCount)
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
	if u, err := FetchCodeforcesUsersInfo([]string{handle}, false); err != nil {
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

func UpdateCodeforcesUserRatingChanges(handle string) error {
	dbConnection := db.GetDBConnection()
	var user types.User
	if result := dbConnection.Where("handle = ?", handle).First(&user); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			if err := UpdateCodeforcesUserInfo(handle); err != nil {
				return err
			}

			if err := dbConnection.Where("handle = ?", handle).First(&user).Error; err != nil {
				return fmt.Errorf("panic err while fetch user: Unexpected brach: %v", err)
			}

		} else {
			return fmt.Errorf("failed to find user %s in DB: %v", handle, result.Error)
		}
	}

	if time.Since(user.RatingChangeUpdateAt) <= 30*time.Minute {
		return nil
	}

	ratingChanges, err := FetchCodeforcesUserRatingChanges(handle)
	if err != nil {
		return err
	}
	user.RatingChanges = *ratingChanges
	if length := len(*ratingChanges); length > 0 {
		user.RatingChangeUpdateAt = time.Now()
	}
	if err := dbConnection.Save(&user).Error; err != nil {
		return err
	}
	return nil
}

func UpdateCodeforcesUserInfo(handle string) error {
	user, err := FetchCodeforcesUsersInfo([]string{handle}, false)
	if err != nil {
		return fmt.Errorf("failed to update cf user: %v", err)
	}
	return db.GetDBConnection().Save(&((*user)[0])).Error
}
