package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/time/rate"

	"github.com/PuerkitoBio/goquery"
	"github.com/YourSuzumiya/ACMBot/app/errs"
	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
)

const (
	baseURL     = "https://atcoder.jp"
	apiBaseURL  = "https://kenkoooo.com/atcoder"
	rateLimit   = 50 * time.Millisecond
	httpTimeout = 30 * time.Second
)

var (
	// 全局限流器，所有 API 请求共享
	apiLimiter = rate.NewLimiter(rate.Every(rateLimit), 1)
)

type Config struct {
	BaseURL     string
	APIBaseURL  string
	RateLimit   time.Duration
	HTTPTimeout time.Duration
}

type AtcoderFetcher struct {
	config  *Config
	client  *colly.Collector
	limiter *rate.Limiter
}

func NewAtcoderFetcher(config *Config) *AtcoderFetcher {
	c := colly.NewCollector(
		colly.AllowedDomains("atcoder.jp"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"),
	)

	return &AtcoderFetcher{
		config:  config,
		client:  c,
		limiter: rate.NewLimiter(rate.Every(config.RateLimit), 1),
	}
}

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

type AtcoderError struct {
	StatusCode int
	Message    string
}

func (e *AtcoderError) Error() string {
	return fmt.Sprintf("atcoder error: status=%d, message=%s", e.StatusCode, e.Message)
}


// API 相关函数
func fetchAPI[T any](suffix string, args map[string]any) (*T, error) {
	requestURL := apiBaseURL + "/" + suffix + "?"
	for k, v := range args {
		requestURL += k + "=" + fmt.Sprint(v) + "&"
	}
	requestURL = requestURL[:len(requestURL)-1]

	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	// 使用全局限流器
	if err := apiLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}

	response, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var res T
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func FetchAtcoderUserSubmissionList(handle string, from int64) (*[]AtcoderUserSubmission, error) {
	return fetchAPI[[]AtcoderUserSubmission]("atcoder-api/v3/user/submissions", map[string]any{
		"user":        handle,
		"from_second": from,
	})
}

func FetchAtcoderContestList() (*[]AtcoderContest, error) {
	return fetchAPI[[]AtcoderContest]("resources/contests.json", nil)
}

// 独立的网页爬虫函数
func FetchAtcoderUser(handle string) (*AtcoderUser, error) {
	logger := log.WithFields(log.Fields{
		"handle": handle,
		"action": "fetch_user",
	})

	logger.Info("Starting fetch user")
	user := &AtcoderUser{Handle: handle}
	var e error

	d := colly.NewCollector(
		colly.AllowedDomains("atcoder.jp"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"),
	)

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

		switch r.StatusCode {
		case http.StatusNotFound:
			e = &errs.ErrHandleNotFound{Handle: handle}
		default:
			e = &AtcoderError{
				StatusCode: r.StatusCode,
				Message:    err.Error(),
			}
		}
		log.WithFields(log.Fields{
			"handle": handle,
			"error":  e,
		}).Info("Failed to fetch Atcoder user")
	})

	url := baseURL + "/users/" + handle
	log.Infof("Visiting: %v", url)
	_ = d.Visit(url)

	logger.Info("Completed fetch user")
	return user, e
}

func atoui(s string) (uint, error) {
	tmp, err := strconv.Atoi(s)
	return uint(tmp), err
}
