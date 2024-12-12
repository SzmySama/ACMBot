package qq

import (
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app"
	"github.com/YourSuzumiya/ACMBot/app/bot"
	myMsg "github.com/YourSuzumiya/ACMBot/app/bot/message"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	zMsg "github.com/wdvxdr1123/ZeroBot/message"
	"strings"
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
			ProtoType: bot.ProtoTypeQQ,
			StepValue: nil,
		},
	}

	for _, opt := range opts {
		opt(res)
	}

	return res
}

func (c *qqContext) GetSender() bot.SenderInfo {
	return bot.SenderInfo{
		ID:       c.zCtx.Event.UserID,
		GroupID:  c.zCtx.Event.GroupID,
		NickName: c.zCtx.Event.Sender.NickName,
	}
}

func (c *qqContext) GetContextType() bot.ProtoType {
	return bot.ProtoTypeQQ
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
			appendFunc = func(myMsg.Node) {
				resultMessage = append(resultMessage, zMsg.CustomNode("", 0, trans(v)))
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

	for commands, task := range bot.CommandMap {
		for _, command := range *commands {
			zero.OnCommand(command).Handle(func(ctx *zero.Ctx) {
				qCtx := newQQContext(withZeroCtx(ctx))
				c := &bot.Context{
					Invoker:   qCtx,
					ProtoType: qCtx.ProtoType,
				}
				err := task(c)
				if err != nil {
					qCtx.Send(myMsg.Message{myMsg.Text(err.Error())})
					qCtx.SendError(err)
				}
			})
		}
	}

	zero.Run(&zeroCfg)
}
