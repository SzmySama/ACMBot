package bot

import (
	"github.com/YourSuzumiya/ACMBot/app"
	"github.com/YourSuzumiya/ACMBot/app/helper"
	"github.com/YourSuzumiya/ACMBot/app/manager"
	"github.com/YourSuzumiya/ACMBot/app/model"
)

var (
	CommandPrefix = app.GetConfig().Bot.CommandPrefix

	MenuText = `以下是功能列表：所有命令都要加上前缀` + "`" + CommandPrefix + "`" + `哦🥰
0. help(或菜单)，输出本消息

1. cf [username]，用于查询codeforces用户的基本信息

2. rating(或rt) [username]，用于查询codeforces用户的rating变化曲线

3. 近期[比赛,atc,nk,lg,cf]，用于查询近期的比赛数据，数据来源于clist.by`
)

type CommandHandler struct {
	Commands []string
	Handler  Task
}

var (
	Commands = []CommandHandler{
		{[]string{"近期比赛"}, raceHandler(manager.GetAllCachedRaces)},
		{[]string{"近期cf"}, raceHandler(manager.GetCachedRacesByResource(model.ResourceCodeforces))},
		{[]string{"近期atc"}, raceHandler(manager.GetCachedRacesByResource(model.ResourceAtcoder))},
		{[]string{"近期nk"}, raceHandler(manager.GetCachedRacesByResource(model.ResourceNowcoder))},
		{[]string{"近期lg"}, raceHandler(manager.GetCachedRacesByResource(model.ResourceLuogu))},

		{[]string{"cf"}, codeforcesProfileHandler},
		{[]string{"rt", "rating"}, codeforcesRatingHandler},

		{[]string{"help", "菜单"}, textHandler(MenuText)},
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

func textHandler(text string) Task {
	return func(ctx *Context) error {
		ctx.StepValue = text
		return sendText(ctx)
	}
}

func bindCodeforcesUserHandler(ctx *Context) error {
	return helper.NewChainContext[Context](ctx).
		Then(getHandlerFromParams).
		Then(bindCodeforcesUser).
		Execute()
}

func qqGroupRankHandler(ctx *Context) error {
	return qqGroupRank(ctx)
}
