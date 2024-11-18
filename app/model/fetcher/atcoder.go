package fetcher

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
)

type AtcoderUser struct {
	Name             string
	Rank             string
	Rating           uint
	IsProvisional    bool   // 分数是否与水平相符
	Dan              string // 段位
	PromotionMessage string // 升段信息
	HighestRating    uint
	RatedMatches     uint
	LastCompeted     string
}

var c = colly.NewCollector(
	colly.AllowedDomains("atcoder.jp"),
)

func FetchAtcoderUser(username string) (*AtcoderUser, error) {
	user := &AtcoderUser{Name: username}
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

	d.OnError(func(r *colly.Response, err error) {
		if r != nil {
			user, e = nil, err
			if r.StatusCode == http.StatusNotFound {
				log.Infof(fmt.Sprintf("User not found: %v", username))
			} else {
				log.Infof(fmt.Sprintf("Failed to fetch Atcoder user: %v", err))
			}
		}
	})

	url := "https://atcoder.jp/users/" + username
	log.Infof("Visiting: %v", url)
	d.Visit(url)

	return user, e
}

func atoui(s string) (uint, error) {
	tmp, err := strconv.Atoi(s)
	return uint(tmp), err
}
