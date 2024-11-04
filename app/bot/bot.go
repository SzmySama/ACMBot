package bot

import (
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app/model/db"
	"github.com/YourSuzumiya/ACMBot/app/model/fetcher"
	"github.com/YourSuzumiya/ACMBot/app/render"
	"github.com/YourSuzumiya/ACMBot/app/utils/config"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"
	"strings"
)

const (
	QueryLimit    = 3
	CommandPrefix = "#"
)

var (
	cfg     = config.GetConfig().WS
	zeroCfg = zero.Config{
		NickName:      []string{"bot"},
		CommandPrefix: CommandPrefix,
		SuperUsers:    []int64{1549992006},
		Driver: []zero.Driver{
			driver.NewWebSocketClient(
				fmt.Sprintf("ws://%s:%d", cfg.Host, cfg.Port),
				cfg.Token),
		},
	}
)

func codeforcesUserProfile(handle string, ctx *zero.Ctx) {
	if err := fetcher.UpdateDBCodeforcesUser(handle); err != nil {
		ctx.Send("获取数据的时候出错惹🥹: " + err.Error())
		return
	}

	var user db.CodeforcesUser

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

func codeforcesUserProfileHandler(ctx *zero.Ctx) {
	handles := strings.Split(ctx.MessageString(), " ")[1:]
	if len(handles) > QueryLimit {
		ctx.Send("发这么多会坏掉的🥰")
		return
	}

	for _, handle := range handles {
		go codeforcesUserProfile(handle, ctx)
	}
}

func processCodeforcesRatingChange(handle string, ctx *zero.Ctx) {
	err := fetcher.UpdateDBCodeforcesUser(handle)
	if err != nil {
		ctx.Send("更新用户`" + handle + "`的数据失败惹🥹：" + err.Error())
		return
	}

	var user db.CodeforcesUser
	if err := db.GetDBConnection().Preload("RatingChanges").Where("handle = ?", handle).First(&user).Error; err != nil {
		ctx.Send("DB Err😰: " + err.Error())
	}

	imgData, err := render.CodeforcesRatingChanges(user.RatingChanges, handle)
	if err != nil {
		ctx.Send(fmt.Sprintf("render err😰: %v", err))
		logrus.Warnf("render err😰: %v", err)
		return
	}
	ctx.Send([]message.MessageSegment{message.ImageBytes(imgData)})
}

func codeforcesRatingChangeHandler(ctx *zero.Ctx) {
	handles := strings.Split(ctx.MessageString(), " ")[1:]
	if len(handles) > QueryLimit {
		ctx.Send("发这么多会坏掉的🥰")
		return
	}

	for _, i := range handles {
		go processCodeforcesRatingChange(i, ctx)
	}
}

func allRaceHandler(ctx *zero.Ctx) {
	race := fetcher.GetStuAcmRaces()
	if race.Err != nil {
		ctx.Send("检查到错误，数据可能并未及时更新，上次更新时间: " + race.LastUpdate.String() + "\nErr: " + race.Err.Error())
	}
	ctx.Send(race.MessageSegments)
}

func codeforcesRaceHandler(ctx *zero.Ctx) {
	race := fetcher.GetCodeforcesRaces()
	if race.Err != nil {
		ctx.Send("检查到错误，数据可能并未及时更新，上次更新时间: " + race.LastUpdate.String() + "\nErr: " + race.Err.Error())
	}
	ctx.Send(race.MessageSegments)
}

func menuHandler(ctx *zero.Ctx) {
	ctx.Send(fmt.Sprintf(""+
		"以下是功能菜单：所有命令都要加上前缀`%s`🥰\n"+
		"1.cf [username]，用于查询codeforces用户的基本信息\n"+
		"2.rating(或rt) [username]，用于查询codeforces用户的rating变化曲线\n"+
		"3.近期比赛，用于查询近期的比赛数据，数据来源于sdutacm.cn\n"+
		"4.近期cf，用于查询近期的codeforces数据，数据来源codeforces.com\n"+
		"项目地址https://github.com/YourSuzumiya/ACMBot，喜欢可以加个Star支持一下\n"+
		"Bot可以直接拉到自己群里用哦",
		CommandPrefix,
	))
}

func init() {
	zero.OnCommand("近期比赛").Handle(allRaceHandler)

	zero.OnCommand("近期cf").Handle(codeforcesRaceHandler)
	zero.OnCommand("rating").Handle(codeforcesRatingChangeHandler)
	zero.OnCommand("rt").Handle(codeforcesRatingChangeHandler)

	zero.OnCommand("cf").Handle(codeforcesUserProfileHandler)

	zero.OnCommand("菜单").Handle(menuHandler)
	zero.OnCommand("help").Handle(menuHandler)
}

func Start() {
	zero.RunAndBlock(&zeroCfg, func() {
		zero.RangeBot(func(_ int64, ctx *zero.Ctx) bool {
			go fetcher.Updater(ctx)
			return false
		})
	})
}
