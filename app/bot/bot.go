package bot

import (
	"fmt"
	"strings"

	"github.com/YourSuzumiya/ACMBot/app/model/db"
	"github.com/YourSuzumiya/ACMBot/app/model/fetcher"
	"github.com/YourSuzumiya/ACMBot/app/model/render"
	"github.com/YourSuzumiya/ACMBot/app/utils/config"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	QueryLimit    = 1
	CommandPrefix = "#"
)

var (
	BotCfg  = config.GetConfig().Bot
	zeroCfg = zero.Config{
		NickName:      BotCfg.NickName,
		CommandPrefix: BotCfg.CommandPrefix,
		SuperUsers:    BotCfg.SuperUsers,
		Driver:        []zero.Driver{},
	}
)

func init() {
	for _, WScfg := range BotCfg.WS {
		zeroCfg.Driver = append(zeroCfg.Driver, driver.NewWebSocketClient(
			fmt.Sprintf("ws://%s:%d", WScfg.Host, WScfg.Port),
			WScfg.Token))
	}
}

func Start() {
	zero.OnCommand("近期比赛").Handle(allRaceHandler)

	zero.OnCommand("近期cf").Handle(codeforcesRaceHandler)
	zero.OnCommand("rating").Handle(codeforcesRatingChangeHandler)
	zero.OnCommand("rt").Handle(codeforcesRatingChangeHandler)

	zero.OnCommand("cf").Handle(codeforcesUserProfileHandler)

	zero.OnCommand("菜单").Handle(menuHandler)
	zero.OnCommand("help").Handle(menuHandler)

	zero.RunAndBlock(&zeroCfg, func() {
		zero.RangeBot(func(_ int64, ctx *zero.Ctx) bool {
			go fetcher.Updater(ctx)
			return false
		})
	})
}

func processCodeforcesUserProfile(handle string, ctx *zero.Ctx) {
	if err := fetcher.UpdateDBCodeforcesUser(handle, ctx); err != nil {
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
	count := 1
	for _, i := range handles {
		if i == "" {
			continue
		}
		if count > QueryLimit {
			ctx.Send("参数太多了🥰，后面的就不查了哦")
			return
		}
		count++
		go processCodeforcesUserProfile(i, ctx)
	}
}

func processCodeforcesRatingChange(handle string, ctx *zero.Ctx) {
	err := fetcher.UpdateDBCodeforcesUser(handle, ctx)
	if err != nil {
		ctx.Send("更新用户`" + handle + "`的数据失败惹🥹：" + err.Error())
		return
	}

	var user db.CodeforcesUser
	if err := db.GetDBConnection().Preload("RatingChanges").Where("handle = ?", handle).First(&user).Error; err != nil {
		ctx.Send("DB Err😰: " + err.Error())
	}

	if len(user.RatingChanges) == 0 {
		ctx.Send("没有找到用户`" + handle + "`的rating记录，赛时加入比赛才计入rating哦")
		return
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
	count := 1
	for _, i := range handles {
		if i == "" {
			continue
		}
		if count > QueryLimit {
			ctx.Send("参数太多了🥰，后面的就不查了哦")
			return
		}
		count++
		go processCodeforcesRatingChange(i, ctx)
	}
}

func allRaceHandler(ctx *zero.Ctx) {
	race, err := fetcher.GetStuAcmRaces()
	if err != nil {
		ctx.Send("检查到错误，数据可能并未及时更新: " + err.Error())
	}
	ctx.Send(race)
}

func codeforcesRaceHandler(ctx *zero.Ctx) {
	race, err := fetcher.GetCodeforcesRaces()
	if err != nil {
		ctx.Send("检查到错误，数据可能并未及时更新: " + err.Error())
	}
	ctx.Send(race)
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
