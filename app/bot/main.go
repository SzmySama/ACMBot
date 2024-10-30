package bot

import (
	"errors"
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
	"gorm.io/gorm"
)

const (
	QueryLimit    = 3
	CommandPrefix = "#"
)

var (
	cfg     = config.GetConfig().RWS
	zeroCfg = zero.Config{
		NickName:      []string{"bot"},
		CommandPrefix: CommandPrefix,
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
	allRace, err := fetcher.GetAndFetchRaces(ctx)
	if err != nil {
		ctx.Send("出错惹🥹: " + err.Error())
	}
	ctx.Send(allRace.AllRacesMessageSegments)
}

func codeforcesUserProfile(handle string, ctx *zero.Ctx) {
	if err := fetcher.UpdateCodeforcesUserSubmissionsAndRating(handle); err != nil {
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
	dbConnection := db.GetDBConnection()
	if err := fetcher.UpdateCodeforcesUserRatingChanges(handle); err != nil {
		ctx.Send(fmt.Sprintf("没有查到%s🥺: %v", handle, err))
		logrus.Warnf("没有查到%s🥺: %v", handle, err)
		return
	}
	var user types.User
	if err := dbConnection.Where("handle = ?", handle).First(&user).Error; err != nil {
		ctx.Send(fmt.Sprintf("DB Err😭: %v", err))
		logrus.Warnf("DB Err😭: %v", err)
		return
	}

	if len(user.RatingChanges) <= 0 {
		ctx.Send(handle + "貌似还没打过比赛")
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
	if len(handles) > QueryLimit {
		ctx.Send("发这么多会坏掉的🥰")
		return
	}

	for _, i := range handles {
		go processCodeforcesRatingChange(i, ctx)
	}
}

func codeforcesRaceHandler(ctx *zero.Ctx) {
	allRace, err := fetcher.GetAndFetchRaces(ctx)
	if err != nil {
		ctx.Send("出错惹🥵: " + err.Error())
	}
	if len(allRace.CodeforcesRacesMessageSegments) > 0 {
		ctx.Send(allRace.CodeforcesRacesMessageSegments)
	} else {
		ctx.Send("近期没有codeforces")
	}
}

func bindCodeforcesHandler(ctx *zero.Ctx) {
	dbConnection := db.GetDBConnection()
	ID := ctx.Event.Sender.ID
	handle := strings.Split(ctx.MessageString(), " ")[1]

	var err error
	var user types.QQUser
	err = dbConnection.FirstOrCreate(&user, types.QQUser{ID: ID}).Error
	if err != nil {
		ctx.Send(fmt.Sprintf("绑定失败😭: %v", err))
		return
	}
	user.CodeforcesHandle = handle
	err = dbConnection.Save(&user).Error
	if err != nil {
		ctx.Send(fmt.Sprintf("绑定失败😭: %v", err))
		return
	}
	ctx.Send("绑定成功")
}

func myCodeforcesHandler(ctx *zero.Ctx) {
	dbConnection := db.GetDBConnection()
	ID := ctx.Event.Sender.ID

	var handle string
	err := dbConnection.Model(&types.QQUser{}).
		Select("CodeforcesHandle").
		Where("id = ?", ID).
		Limit(1).
		Scan(&handle).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Send("没有查询到你的绑定信息，快来绑定吧🥰")
			return
		}
		ctx.Send(fmt.Sprintf("DB Err😰: %v", err))
		return
	}

	codeforcesUserProfile(handle, ctx)
}

func myRatingHandler(ctx *zero.Ctx) {
	dbConnection := db.GetDBConnection()
	ID := ctx.Event.Sender.ID

	var handle string
	err := dbConnection.Model(&types.QQUser{}).
		Select("CodeforcesHandle").
		Where("id = ?", ID).
		Limit(1).
		Scan(&handle).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Send("没有查询到你的绑定信息，快来绑定吧🥰")
			return
		}
		ctx.Send(fmt.Sprintf("DB Err😰: %v", err))
		return
	}

	processCodeforcesRatingChange(handle, ctx)
}

func menuHandler(ctx *zero.Ctx) {
	ctx.Send(fmt.Sprintf(""+
		"以下是功能菜单：所有命令都要加上前缀`%s`🥰\n"+
		"1.cf [username]，用于查询codeforces用户的基本信息\n"+
		"2.rating(或rt) [username]，用于查询codeforces用户的rating变化曲线\n"+
		"3.近期比赛，用于查询近期的比赛数据，数据来源于sdutacm.cn\n"+
		"4.近期cf，用于查询近期的codeforces数据，数据来源codeforces.com\n"+
		"项目地址https://github.com/SzmySama/ACMBot，喜欢可以加个Star支持一下\n"+
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

	zero.OnCommand("绑定cf").Handle(bindCodeforcesHandler)
	zero.OnCommand("我的cf").Handle(myCodeforcesHandler)
	zero.OnCommand("我的rt").Handle(myRatingHandler)

	zero.OnCommand("菜单").Handle(menuHandler)
}

func Start() {
	zero.RunAndBlock(&zeroCfg, nil)
}
