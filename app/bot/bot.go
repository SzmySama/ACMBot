package bot

import (
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app/model/manager"
	"strings"

	"github.com/YourSuzumiya/ACMBot/app/utils/config"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	QueryLimit    = 3
	CommandPrefix = "#"
)

var (
	Cfg     = config.GetConfig().Bot
	zeroCfg = zero.Config{
		NickName:      Cfg.NickName,
		CommandPrefix: Cfg.CommandPrefix,
		SuperUsers:    Cfg.SuperUsers,
		Driver:        []zero.Driver{},
	}
)

func init() {
	for _, cfg := range Cfg.WS {
		zeroCfg.Driver = append(zeroCfg.Driver, driver.NewWebSocketClient(
			fmt.Sprintf("ws://%s:%d", cfg.Host, cfg.Port),
			cfg.Token))
	}
}

func Start() {
	zero.OnCommand("近期比赛").Handle(allRaceHandler)

	zero.OnCommand("近期cf").Handle(codeforcesRaceHandler)
	zero.OnCommand("rating").Handle(codeforcesRatingChangeHandler)
	zero.OnCommand("rt").Handle(codeforcesRatingChangeHandler)

	zero.OnCommand("cf").Handle(codeforcesUserProfileV2Handler)

	zero.OnCommand("bind").Handle(bindCodeforcesIDHandler)
	zero.OnCommand("菜单").Handle(menuHandler)
	zero.OnCommand("help").Handle(menuHandler)

	go manager.RaceUpdater()

	zero.RunAndBlock(&zeroCfg, nil)
}

func processCodeforcesUserProfile(handle string, ctx *zero.Ctx) {
	user, err := manager.GetUpdatedCodeforcesUser(handle)
	if err != nil {
		ctx.Send(err.Error())
		return
	}
	image, err := user.ToRenderProfileV1().ToImage()
	if err != nil {
		ctx.Send(err.Error())
	}
	ctx.Send([]message.MessageSegment{message.ImageBytes(image)})
}

func codeforcesUserProfileHandler(ctx *zero.Ctx) {
	handles := strings.Split(ctx.MessageString(), " ")[1:]
	if len(handles) == 0 {
		ctx.Send("没听到你要查谁呢，再说一遍吧？")
		return
	}

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

func processCodeforcesUserProfileV2(handle string, ctx *zero.Ctx) {
	user, err := manager.GetUpdatedCodeforcesUser(handle)
	if err != nil {
		ctx.Send(err.Error())
		return
	}
	image, err := user.ToRenderProfileV2().ToImage()
	if err != nil {
		ctx.Send(err.Error())
	}
	ctx.Send([]message.MessageSegment{message.ImageBytes(image)})
}

func codeforcesUserProfileV2Handler(ctx *zero.Ctx) {
	handles := strings.Split(ctx.MessageString(), " ")[1:]
	if len(handles) == 0 {
		ctx.Send("没听到你要查谁呢，再说一遍吧？")
		return
	}

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
		go processCodeforcesUserProfileV2(i, ctx)
	}
}

func processCodeforcesRatingChange(handle string, ctx *zero.Ctx) {
	user, err := manager.GetUpdatedCodeforcesUser(handle)
	if err != nil {
		ctx.Send(err.Error())
		return
	}
	image, err := user.ToRenderRatingChanges().ToImage()
	if err != nil {
		ctx.Send(err.Error())
		return
	}
	ctx.Send([]message.MessageSegment{message.ImageBytes(image)})
}

func codeforcesRatingChangeHandler(ctx *zero.Ctx) {
	handles := strings.Split(ctx.MessageString(), " ")[1:]
	if len(handles) == 0 {
		ctx.Send("没听到你要查谁呢，再说一遍吧？")
		return
	}

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

func bindCodeforcesIDHandler(ctx *zero.Ctx) {
	codeforcesID := strings.Split(ctx.MessageString(), " ")[1:]
	if len(codeforcesID) > 1 {
		ctx.Send("[CQ:at,qq=" + fmt.Sprint(ctx.Event.UserID) + "] " +
			"一个QQ号只能绑定一个codeforces账号哦")
		return
	} else if len(codeforcesID) == 0 {
		ctx.Send("[CQ:at,qq=" + fmt.Sprint(ctx.Event.UserID) + "] " +
			"没听到你要绑定到哪个账号呢，请再说一遍吧")
		return
	}
	err := manager.BindQQAndCodeforcesHandler(uint(ctx.Event.UserID), uint(ctx.Event.GroupID), codeforcesID[0])
	if err != nil {
		ctx.Send("[CQ:at,qq=" + fmt.Sprint(ctx.Event.UserID) + "] " + err.Error())
		return
	}
	ctx.Send("[CQ:at,qq=" + fmt.Sprint(ctx.Event.UserID) + "] " +
		"成功绑定 " + ctx.Event.Sender.NickName + " -> " + codeforcesID[0])
}

func allRaceHandler(ctx *zero.Ctx) {
	race, err := manager.GetStuACMRaces()
	if err != nil {
		ctx.Send("检查到错误，数据可能并未及时更新: " + err.Error())
	}
	ctx.Send(race)
}

func codeforcesRaceHandler(ctx *zero.Ctx) {
	race, err := manager.GetCodeforcesRaces()
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
		"5.bind [username]，用于绑定codeforces账号\n"+
		"6.rank，用于查询群内codeforces排行榜\n"+
		"项目地址https://github.com/YourSuzumiya/ACMBot，喜欢可以加个Star支持一下\n"+
		"Bot可以直接拉到自己群里用哦",
		CommandPrefix,
	))
}
