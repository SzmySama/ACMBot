package bot

import (
	"github.com/YourSuzumiya/ACMBot/app"
	"github.com/YourSuzumiya/ACMBot/app/helper"
	"github.com/YourSuzumiya/ACMBot/app/manager"
	"github.com/YourSuzumiya/ACMBot/app/model"
)

var (
	CommandPrefix = app.GetConfig().Bot.CommandPrefix

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
	CommandMap = map[string]Task{

		"近期比赛":  raceHandler(manager.GetStuACMRaces),
		"近期cf":  raceHandler(manager.GetCodeforcesRaces),
		"近期atc": raceHandler(manager.GetAtCoderRaces),
		"近期nk":  raceHandler(manager.GetNowCoderRaces),
		"近期lg":  raceHandler(manager.GetLuoguRaces),

		"cf": codeforcesProfileHandler,
		"rt": codeforcesRatingHandler,
	}
)

func codeforcesProfileHandler(ctx *Context) error {
	return helper.
		NewChainContext[Context](ctx).
		Then(getHandlerFromParams).
		Then(getCodeforcesUserByHandle).
		Then(getRenderedCodeforcesUserProfile).
		Then(sendPicture).
		Execute()
}

func codeforcesRatingHandler(ctx *Context) error {
	return helper.
		NewChainContext[Context](ctx).
		Then(getHandlerFromParams).
		Then(getCodeforcesUserByHandle).
		Then(getRenderedCodeforcesRatingChanges).
		Then(sendPicture).
		Execute()
}

func raceHandler(provider model.RaceProvider) Task {
	return func(ctx *Context) error {
		ctx.StepValue = provider
		return helper.
			NewChainContext[Context](ctx).
			Then(getRaceFromProvider).
			Then(sendRace).
			Execute()
	}
}
