package bot

import (
	"fmt"
	"strings"

	"github.com/SzmySama/ACMBot/app/fetcher"
	"github.com/SzmySama/ACMBot/app/render"
	"github.com/SzmySama/ACMBot/app/types"
	"github.com/SzmySama/ACMBot/app/utils/config"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"
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
		ctx.Send("出错惹🥵: " + err.Error())
	}
	var result message.Message
	for _, v := range allRace {
		result = append(result, message.CustomNode("", 0, v.String()))
	}
	ctx.Send(result)
}

func codeforcesUserProfileHandler(ctx *zero.Ctx) {
	handles := strings.Split(ctx.MessageString(), " ")[1:]
	users, err := fetcher.FetchCodeforcesUsersInfo(handles, false)
	if err != nil {
		ctx.Send("没有找到这位用户🥵: " + err.Error())
		return
	}
	logrus.Infof("%v", users)
	geneAndSend := func(user types.User) {
		data, err := render.CodeforcesUserProfile(render.CodeforcesUserProfileData{
			User:  user,
			Level: render.ConvertRatingToLevel(user.Rating),
		})
		if err != nil {
			ctx.Send("正在生成" + user.Handle + "的卡片，但是出错惹🥵: " + err.Error())
		}
		ctx.Send([]message.MessageSegment{message.ImageBytes(data)})
	}
	for _, user := range *users {
		go geneAndSend(user)
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

	zero.OnCommand("cf").Handle(codeforcesUserProfileHandler)

}

func Start() {
	zero.RunAndBlock(&zeroCfg, nil)
}
