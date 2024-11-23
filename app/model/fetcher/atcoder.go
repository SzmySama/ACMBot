package fetcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/time/rate"

	"github.com/PuerkitoBio/goquery"
	"github.com/YourSuzumiya/ACMBot/app/model/errs"
	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
)

type AtcoderUser struct {
	Handle           string // Atcoder用户名
	Avatar           string // 头像URL
	Rank             string
	Rating           uint
	IsProvisional    bool   // 分数是否与水平相符
	Dan              string // 段位
	PromotionMessage string // 升段信息
	HighestRating    uint
	RatedMatches     uint
	LastCompeted     string
}

type AtcoderUserSubmission struct {
	SubmissionId   uint    `json:"id"`
	SubmissionTime int64   `json:"epoch_second"` // 提交时间（UNIX时间戳）
	ProblemId      string  `json:"problem_id"`
	ContestId      string  `json:"contest_id"`
	Handle         string  `json:"user_id"`
	Language       string  `json:"language"`
	Point          float32 `json:"point"`
	Length         uint    `json:"length"`
	Status         string  `json:"result"` // 提交状态
	ExecutionTime  int     `json:"execution_time"`
}

type AtcoderContest struct {
	Id             string `json:"id"`                 // 比赛ID
	StartTime      int64  `json:"start_epoch_second"` // 开始时间（UNIX时间戳）
	DurationSecond uint   `json:"duration_second"`    // 持续时间（秒）
	Title          string `json:"title"`              // 完整标题
	RateChange     string `json:"rate_change"`        // Rated分数范围
}

var (
	c = colly.NewCollector(
		colly.AllowedDomains("atcoder.jp"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"),
	)
	atcoderLimiter = rate.NewLimiter(rate.Every(2*time.Second), 1)
)

// handle: Atcoder用户名
func FetchAtcoderUser(handle string) (*AtcoderUser, error) {
	user := &AtcoderUser{Handle: handle}
	var e error = nil

	d := c.Clone()
	d.OnHTML("tr", func(h *colly.HTMLElement) {
		var err error
		switch h.ChildText("th") {
		case "Rank":
			user.Rank = h.ChildText("td")
		case "Rating":
			td := h.DOM.Find("td")
			user.Rating, err = atoui(td.Find(`span[class^="user-"]`).Text())
			if err != nil {
				log.Infof("Failed to convert rating to uint: %v", err)
				user = nil
				e = fmt.Errorf("Parse error: %v", err)
			}
			user.IsProvisional = td.Find("span").HasClass("bold small")
		case "Highest Rating":
			h.DOM.Find("td").Find("span").Each(func(i int, s *goquery.Selection) {
				switch i {
				case 0:
					user.HighestRating, err = atoui(s.Text())
					if err != nil {
						log.Infof("Failed to convert highest rating to uint: %v", err)
						user = nil
						e = fmt.Errorf("Parse error: %v", err)
					}
				case 2:
					user.Dan = s.Text()
				case 3:
					user.PromotionMessage = s.Text()
				}
			})
		case "Rated Matches":
			user.RatedMatches, err = atoui(h.ChildText("td"))
			if err != nil {
				log.Infof("Failed to convert rated matches to uint: %v", err)
				user = nil
				e = fmt.Errorf("Parse error: %v", err)
			}
		case "Last Competed":
			user.LastCompeted = h.ChildText("td")
		}
	})

	d.OnHTML("img.avatar", func(h *colly.HTMLElement) {
		user.Avatar = h.Attr("src")
	})

	d.OnError(func(r *colly.Response, err error) {
		if r == nil {
			return
		}

		user, e = nil, err
		if r.StatusCode == http.StatusNotFound {
			e = errs.ErrHandleNotFound
			log.Infof("Handle not found: %v", handle)
			return
		}
		log.Infof("Failed to fetch Atcoder user: %v", err)
	})

	url := "https://atcoder.jp/users/" + handle
	log.Infof("Visiting: %v", url)

	// Depress Warning
	_ = d.Visit(url)

	return user, e
}

func fetchAtcoderAPI[T any](suffix string, args map[string]any) (*T, error) {
	requestURL := "https://kenkoooo.com/atcoder/" + suffix + "?"
	for k, v := range args {
		requestURL += k + "=" + fmt.Sprint(v) + "&"
	}

	requestURL = requestURL[:len(requestURL)-1]
	log.Infof("Visiting: %v", requestURL)

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()
	if err := atcoderLimiter.Wait(ctx); err != nil {
		log.Infof("Timeout or cancelled while waiting: %v", err)
		return nil, err
	}

	response, err := http.Get(requestURL)
	if err != nil && errors.Is(err, io.EOF) {
		log.Infof("Failed to fetch Atcoder API: %v", err)
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Infof("Failed to read response body: %v", err)
		return nil, err
	}

	if err := response.Body.Close(); err != nil {
		log.Infof("Failed to close response body: %v", err)
		return nil, err
	}

	var res T
	if err := json.Unmarshal(body, &res); err != nil {
		log.Infof("Failed to unmarshal response body: %v\n %v", string(body), err)
		return nil, err
	}

	return &res, nil
}

// 获取名为handle的用户，时间从from开始的最多500条提交，from为UNIX时间戳
func FetchAtcoderUserSubmissionList(handle string, from int64) (*[]AtcoderUserSubmission, error) {
	return fetchAtcoderAPI[[]AtcoderUserSubmission]("atcoder-api/v3/user/submissions", map[string]any{
		"user":        handle,
		"from_second": from,
	})
}

// 获取Atcoder比赛列表
func FetchAtcoderContestList() (*[]AtcoderContest, error) {
	return fetchAtcoderAPI[[]AtcoderContest]("resources/contests.json", nil)
}

func atoui(s string) (uint, error) {
	tmp, err := strconv.Atoi(s)
	return uint(tmp), err
}
