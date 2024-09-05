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

	"github.com/SzmySama/ACMBot/app/types"
	"github.com/SzmySama/ACMBot/app/utils/config"
	log "github.com/sirupsen/logrus"
)

func fetchCodeforcesAPI(apiMethod string, args map[string]any) ([]map[string]any, error) {
	apiURL := "https://codeforces.com/api/"
	cfg := config.GetConfig().Codeforces

	args["apiKey"] = cfg.Key
	args["time"] = strconv.Itoa(int(time.Now().Unix()))

	var sortedArgs []string
	for k, v := range args {
		sortedArgs = append(sortedArgs, fmt.Sprintf("%s=%s", k, v))
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

	var res codeforcesResponse
	if err := json.Unmarshal([]byte(body), &res); err != nil {
		return nil, err
	}

	if res.Status != "OK" {
		log.Warnf("Status is not OK")
		return nil, fmt.Errorf(res.Comment)
	}

	return res.Result, nil
}

func FetchCodeforcesUsersInfo(handles []string) (user []types.User, err error) {
	var result []map[string]any
	result, err = fetchCodeforcesAPI("user.info", map[string]any{
		"handles":              strings.Join(handles, ";"),
		"checkHistoricHandles": "false",
	})
	if err != nil {
		return
	}
	var json_result []byte

	for _, v := range result {
		var currect_user types.User

		json_result, err = json.Marshal(v)
		if err != nil {
			return
		}
		log.Infof(string(json_result))
		if err = json.Unmarshal(json_result, &currect_user); err != nil {
			return
		}
		user = append(user, currect_user)
	}
	return
}

// func FetchCodeforcesUserSubmissionsUntil(handle string, until time.Time) ([]types.Submission, error) {
// 	// 获取的提交记录是按照时间排序的，靠近现在的排在前面

// 	const SUBMISSION_COUNT = 1000
// 	var result []types.Submission
// 	var lastTime time.Time

// 	fetchSubmission := func(from int) ([]types.Submission, error) {
// 		res, err := fetchCodeforcesAPI("user.status", map[string]any{
// 			"handle": handle,
// 			"from":   from,
// 			"count":  SUBMISSION_COUNT,
// 		})
// 		if err != nil {
// 			return nil, err
// 		}
// 		return res, err
// 	}

// 	for {

// 	}
// }
