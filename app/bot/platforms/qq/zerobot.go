package qq

import (
	"errors"
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app"
	"github.com/YourSuzumiya/ACMBot/app/bot"
	myMsg "github.com/YourSuzumiya/ACMBot/app/bot/message"
	"github.com/YourSuzumiya/ACMBot/app/errs"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	zMsg "github.com/wdvxdr1123/ZeroBot/message"
	"strings"
	"time"
)

type qqContext struct {
	bot.Context
	zCtx *zero.Ctx
}

type ctxOption func(*qqContext)

func withZeroCtx(zCtx *zero.Ctx) ctxOption {
	return func(ctx *qqContext) {
		ctx.zCtx = zCtx
	}
}

func newQQContext(opts ...ctxOption) *qqContext {
	res := &qqContext{
		Context: bot.Context{
			Platform:  bot.PlatformQQ,
			StepValue: nil,
		},
	}

	for _, opt := range opts {
		opt(res)
	}

	return res
}

func (c *qqContext) GetCallerInfo() bot.CallerInfo {
	result := bot.CallerInfo{
		ID:       c.zCtx.Event.UserID,
		NickName: c.zCtx.Event.Sender.NickName,
	}

	gid := c.zCtx.Event.GroupID

	if gid != 0 {
		gInfo := c.zCtx.GetGroupInfo(gid, false)
		result.Group = bot.GroupInfo{
			ID:          gid,
			Name:        gInfo.Name,
			MemberCount: gInfo.MemberCount,
		}
	}

	return result
}

func (c *qqContext) GetContextType() bot.Platform {
	return bot.PlatformQQ
}

func (c *qqContext) Send(msg myMsg.Message) {
	c.zCtx.Send(msgToZeroMsg(msg))
}

func (c *qqContext) SendError(err error) {
	for _, user := range zeroCfg.SuperUsers {
		c.zCtx.SendPrivateMessage(user, err.Error())
	}
}

func (c *qqContext) Params() myMsg.Message {
	argStr := c.zCtx.State["args"].(string)
	var res myMsg.Message
	for _, s := range strings.Fields(argStr) {
		res = append(res, myMsg.Text(s))
	}
	return res
}

func trans(node myMsg.Node) zMsg.MessageSegment {
	switch node.MessageType {
	case myMsg.TypeText:
		return zMsg.Text(node.Text_())
	case myMsg.TypeImageBytes:
		return zMsg.ImageBytes(node.ImageBytes_())
	case myMsg.TypeAt:
		return zMsg.At(node.At_())
	case myMsg.TypeMixNode:
		return trans(node.MixNode_())
	default:
		return zMsg.Text("Unknown message type")
	}
}

func msgToZeroMsg(msg myMsg.Message) zMsg.Message {
	if len(msg) == 0 {
		return zMsg.Message{}
	}

	resultMessage := make(zMsg.Message, 0, len(msg))

	appendFunc := func(n myMsg.Node) {
		resultMessage = append(resultMessage, trans(n))
	}

	for _, v := range msg {
		if v.MessageType == myMsg.TypeMixNode {
			appendFunc = func(n myMsg.Node) {
				resultMessage = append(resultMessage, zMsg.CustomNode("", 0, zMsg.Message{trans(n)}))
			}
			break
		}
	}

	for _, v := range msg {
		appendFunc(v)
	}

	return resultMessage
}

var (
	zeroCfg zero.Config
)

// TODO: 把配置转移到bot层级
func init() {
	cfg := app.GetConfig().Bot
	zeroCfg = zero.Config{
		NickName:      cfg.NickName,
		CommandPrefix: bot.CommandPrefix,
		SuperUsers:    cfg.SuperUsers,
		Driver:        []zero.Driver{},
	}

	for _, cfg := range cfg.WS {
		zeroCfg.Driver = append(zeroCfg.Driver, driver.NewWebSocketClient(
			fmt.Sprintf("ws://%s:%d", cfg.Host, cfg.Port),
			cfg.Token))
	}

	zero.OnRequest(func(ctx *zero.Ctx) bool {
		return ctx.Event.RequestType == "group"
	}).Handle(func(ctx *zero.Ctx) {
		ctx.SetGroupAddRequest(ctx.Event.Flag, ctx.Event.SubType, true, "")
		for _, user := range cfg.SuperUsers {
			ctx.SendPrivateMessage(user, fmt.Sprintf("已自动同意加群邀请: %d", ctx.Event.GroupID))
		}
		ctx.SendPrivateMessage(ctx.Event.UserID, fmt.Sprintf("已自动同意加群邀请: %d", ctx.Event.GroupID))
		ctx.SendPrivateMessage(ctx.Event.UserID, "要去新地方了呢~\n如果大家不知道如何使用，可以用`#help`, `#菜单`呼出功能列表哦")
	})

	zero.OnRequest(func(ctx *zero.Ctx) bool {
		return ctx.Event.RequestType == "friend"
	}).Handle(func(ctx *zero.Ctx) {
		ctx.SetFriendAddRequest(ctx.Event.Flag, true, "")
		for _, user := range cfg.SuperUsers {
			ctx.SendPrivateMessage(user, fmt.Sprintf("已自动同意好友请求: %d", ctx.Event.UserID))
		}
		go func() {
			time.Sleep(5 * time.Second)
			ctx.SendPrivateMessage(ctx.Event.UserID, "很高兴认识你~ \n"+bot.MenuText)
		}()
	})

	for _, command := range bot.Commands {
		commands := command.Commands
		handler := command.Handler
		zeroHandler := func(ctx *zero.Ctx) {
			qCtx := newQQContext(withZeroCtx(ctx))
			c := &bot.Context{
				Invoker:  qCtx,
				Platform: qCtx.Platform,
			}
			err := handler(c)
			if err == nil {
				return
			}
			qCtx.Send(myMsg.Message{myMsg.Text(err.Error())})
			var internalError errs.InternalError
			if errors.As(err, &internalError) {
				qCtx.SendError(err)
			}
		}

		for _, command := range commands {
			zero.OnCommand(command).Handle(zeroHandler)
		}
	}
	zero.Run(&zeroCfg)
}
