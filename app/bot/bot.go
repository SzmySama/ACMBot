package bot

import (
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app/manager"
	"github.com/YourSuzumiya/ACMBot/app/model"
	"github.com/sirupsen/logrus"
)

const (
	QueryLimit    = 3
	CommandPrefix = "#"

	menuText = `以下是功能菜单：所有命令都要加上前缀` + CommandPrefix + `🥰
1.cf [username]，用于查询codeforces用户的基本信息
2.rating(或rt) [username]，用于查询codeforces用户的rating变化曲线
3.近期比赛，用于查询近期的比赛数据，数据来源于sdutacm.cn
4.近期cf，用于查询近期的codeforces数据，数据来源codeforces.com
5.近期atc，用于查询近期的atcoder数据，数据来源sdutacm.com
6.近期nk，用于查询近期的牛客数据，数据来源sdutacm.com
7.近期lg，用于查询近期的洛谷数据，数据来源sdutacm.com
项目地址https://github.com/YourSuzumiya/ACMBot，喜欢可以加个Star支持一下
Bot可以直接拉到自己群里用哦`
)

var (
	CommandMap = map[string]Handler{
		"test": testHandler,

		"近期比赛":  raceHandler(manager.GetStuACMRaces, "获取近期比赛"),
		"近期cf":  raceHandler(manager.GetCodeforcesRaces, "获取近期cf"),
		"近期atc": raceHandler(manager.GetAtCoderRaces, "获取近期atc"),
		"近期nk":  raceHandler(manager.GetNowCoderRaces, "获取近期nk"),
		"近期lg":  raceHandler(manager.GetLuoguRaces, "获取近期lg"),

		"cf": codeforcesProfileHandler(),
		"rt": codeforcesRatingChangeHandler(),

		"menu": func(ctx Context) {
			ctx.Send(Message{menuText})
		},
	}
)

func testHandler(ctx Context) {
	logrus.Info(ctx.Params())
}

func handlerTemplate[T any](provider func() (T, error), msgGenerator func(T) Message, hint string) Handler {
	return func(ctx Context) {
		result, err := provider()
		if err != nil {
			e := fmt.Errorf("萝卜子在`%s`时遇到了困难: %w", hint, err)
			ctx.SendError(e)
		}
		ctx.Send(msgGenerator(result))
	}
}

// 处理图片消息
func picHandler(provider model.PicProvider, hint string) Handler {
	return handlerTemplate[[]byte](provider, func(picBytes []byte) Message { return Message{picBytes} }, hint)
}

func raceHandler(provider model.RaceProvider, hint string) Handler {
	return handlerTemplate[[]model.Race](provider, func(races []model.Race) Message {
		var msg Message
		for _, race := range races {
			msg = append(msg, race.String())
		}
		return msg
	}, hint)
}

func codeforcesUserHandlerTemplate(processor func(user *manager.CodeforcesUser) model.PicProvider, hint string) Handler {
	return func(ctx Context) {
		params := ctx.Params()
		if len(params) != 1 {
			ctx.Send(Message{"萝卜子温馨提醒: 在" + hint + "时一次只能查询一个用户哦，再问我一次吧"})
			return
		}
		handle, ok := params[0].(string)
		if !ok {
			ctx.Send(Message{"(吓)你发了什么给萝卜子？？"})
			return
		}
		user, err := manager.GetUpdatedCodeforcesUser(handle)
		if err != nil {
			ctx.Send(Message{fmt.Errorf("获取用户信息失败re: %e", err)})
		}
		picHandler(processor(user), hint)
	}
}

func codeforcesProfileHandler() Handler {
	return codeforcesUserHandlerTemplate(func(user *manager.CodeforcesUser) model.PicProvider { return user.ToRenderProfileV2().ToImage }, "查询user profile")
}

func codeforcesRatingChangeHandler() Handler {
	return codeforcesUserHandlerTemplate(func(user *manager.CodeforcesUser) model.PicProvider { return user.ToRenderRatingChanges().ToImage }, "查询user rating曲线图")
}
