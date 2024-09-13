package bot

import (
	"fmt"
	"strings"

	"github.com/SzmySama/ACMBot/app/fetcher"
	"github.com/SzmySama/ACMBot/app/model/db"
	"github.com/SzmySama/ACMBot/app/render"
	"github.com/SzmySama/ACMBot/app/types"
	"github.com/SzmySama/ACMBot/app/utils/config"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	QUERY_LIMIT = 3
)

var (
	cfg     = config.GetConfig().RWS
	zeroCfg = zero.Config{
		NickName:      []string{"bot"},
		CommandPrefix: "#",
		SuperUsers:    []int64{1549992006},
		Driver: []zero.Driver{
			driver.NewWebSocketServer(
				int(cfg.ChannelSize),
				fmt.Sprintf("ws://%s:%d/onebot", cfg.Host, cfg.Port),
				cfg.Token),
		},
	}
)

func allRaceHandler(ctx *zero.Ctx) {
	allRace, err := fetcher.GetAllRaces()
	if err != nil {
		ctx.Send("出错惹🥹: " + err.Error())
	}
	var result message.Message
	for _, v := range allRace {
		result = append(result, message.CustomNode("", 0, v.String()))
	}
	ctx.Send(result)
}

func codeforcesUserProfileHandler(ctx *zero.Ctx) {
	handles := strings.Split(ctx.MessageString(), " ")[1:]
	if len(handles) > QUERY_LIMIT {
		ctx.Send("发这么多会坏掉的🥰")
		return
	}
	geneAndSend := func(handle string) {
		if err := fetcher.UpdateCodeforcesUserSubmissions(handle); err != nil {
			ctx.Send("获取数据的时候出错惹🥹: " + err.Error())
			return
		}

		var user types.User

		if err := db.GetDBConnection().Where("handle = ?", handle).First(&user).Error; err != nil {
			ctx.Send(fmt.Sprintf("DB Err😭: %v", err))
			return
		}

		data, err := render.CodeforcesUserProfile(user)
		if err != nil {
			ctx.Send("正在生成" + user.Handle + "的卡片，但是出错惹🥵: " + err.Error())
			return
		}
		ctx.Send([]message.MessageSegment{message.ImageBytes(data)})
	}
	for _, handle := range handles {
		go geneAndSend(handle)
	}
}

func codeforcesRatingChangeHandler(ctx *zero.Ctx) {
	handles := strings.Split(ctx.MessageString(), " ")[1:]
	if len(handles) > QUERY_LIMIT {
		ctx.Send("发这么多会坏掉的🥰")
		return
	}
	db := db.GetDBConnection()
	genAndSend := func(handle string) {
		if err := fetcher.UpdateCodeforcesUserRatingChanges(handle); err != nil {
			ctx.Send(fmt.Sprintf("没有查到%s🥺: %v", handle, err))
			logrus.Warnf("没有查到%s🥺: %v", handle, err)
			return
		}
		var user types.User
		if err := db.Where("handle = ?", handle).First(&user).Error; err != nil {
			ctx.Send(fmt.Sprintf("DB Err😭: %v", err))
			logrus.Warnf("DB Err😭: %v", err)
			return
		}

		if len(user.RatingChanges) <= 0 {
			ctx.Send(handle + "貌似还没打过比赛")
			return
		}

		img_data, err := render.CodeforcesRatingChanges(user.RatingChanges, handle)
		if err != nil {
			ctx.Send(fmt.Sprintf("render err😰: %v", err))
			logrus.Warnf("render err😰: %v", err)
			return
		}
		ctx.Send([]message.MessageSegment{message.ImageBytes(img_data)})
	}

	for _, i := range handles {
		genAndSend(i)
	}
}

func codeforcesRaceHandler(ctx *zero.Ctx) {
	allRace, err := fetcher.GetAllRaces()
	if err != nil {
		ctx.Send("出错惹🥵: " + err.Error())
	}
	var result message.Message
	for _, v := range allRace {
		if v.Source == "Codeforces" {
			result = append(result, message.CustomNode("", 0, v.String()))
		}
	}
	ctx.Send(result)
}

func init() {
	zero.OnCommand("近期比赛").Handle(allRaceHandler)
	zero.OnCommand("近期cf").Handle(codeforcesRaceHandler)
	zero.OnCommand("rating").Handle(codeforcesRatingChangeHandler)

	zero.OnCommand("cf").Handle(codeforcesUserProfileHandler)

}

func Start() {
	zero.RunAndBlock(&zeroCfg, nil)
}
