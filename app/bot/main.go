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
	ctx.Send(allRace.AllRacesMessageSegments)
}

func process_CodeforcesUserProfile(handle string, ctx *zero.Ctx) {
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

func codeforcesUserProfileHandler(ctx *zero.Ctx) {
	handles := strings.Split(ctx.MessageString(), " ")[1:]
	if len(handles) > QUERY_LIMIT {
		ctx.Send("发这么多会坏掉的🥰")
		return
	}

	for _, handle := range handles {
		go process_CodeforcesUserProfile(handle, ctx)
	}
}

func process_CodeforcesRatingChange(handle string, ctx *zero.Ctx) {
	db := db.GetDBConnection()
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

func codeforcesRatingChangeHandler(ctx *zero.Ctx) {
	handles := strings.Split(ctx.MessageString(), " ")[1:]
	if len(handles) > QUERY_LIMIT {
		ctx.Send("发这么多会坏掉的🥰")
		return
	}

	for _, i := range handles {
		go process_CodeforcesRatingChange(i, ctx)
	}
}

func codeforcesRaceHandler(ctx *zero.Ctx) {
	allRace, err := fetcher.GetAllRaces()
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
	db := db.GetDBConnection()
	ID := ctx.Event.Sender.ID
	handle := strings.Split(ctx.MessageString(), " ")[1]

	var err error
	var user types.QQUser
	err = db.FirstOrCreate(&user, types.QQUser{ID: ID}).Error
	if err != nil {
		ctx.Send(fmt.Sprintf("绑定失败😭: %v", err))
		return
	}
	user.CodeforcesHandle = handle
	err = db.Save(&user).Error
	if err != nil {
		ctx.Send(fmt.Sprintf("绑定失败😭: %v", err))
		return
	}
	ctx.Send("绑定成功")
}

func myCodeforcesHandler(ctx *zero.Ctx) {
	db := db.GetDBConnection()
	ID := ctx.Event.Sender.ID

	var handle string
	err := db.Model(&types.QQUser{}).
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

	process_CodeforcesUserProfile(handle, ctx)
}

func myRatingHandler(ctx *zero.Ctx) {
	db := db.GetDBConnection()
	ID := ctx.Event.Sender.ID

	var handle string
	err := db.Model(&types.QQUser{}).
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

	process_CodeforcesRatingChange(handle, ctx)
}

func init() {
	zero.OnCommand("近期比赛").Handle(allRaceHandler)

	zero.OnCommand("近期cf").Handle(codeforcesRaceHandler)
	zero.OnCommand("rating").Handle(codeforcesRatingChangeHandler)
	zero.OnCommand("cf").Handle(codeforcesUserProfileHandler)

	zero.OnCommand("绑定cf").Handle(bindCodeforcesHandler)
	zero.OnCommand("我的cf").Handle(myCodeforcesHandler)
	zero.OnCommand("我的rt").Handle(myRatingHandler)

}

func Start() {
	zero.RunAndBlock(&zeroCfg, nil)
}
