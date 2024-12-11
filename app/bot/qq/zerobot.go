package qq

import (
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app"
	"github.com/YourSuzumiya/ACMBot/app/bot"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"
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

func (c *qqContext) Send(msg bot.Message) {
	c.zCtx.Send(msgToZeroMsg(msg))
}

func (c *qqContext) SendError(err error) {
	for _, user := range zeroCfg.SuperUsers {
		c.zCtx.SendPrivateMessage(user, err.Error())
	}
}

func (c *qqContext) Params() bot.Message {
	argStr := c.zCtx.State["args"].(string)
	var res bot.Message
	for _, s := range strings.Fields(argStr) {
		res = append(res, s)
	}
	return res
}

func msgToZeroMsg(msg bot.Message) message.Message {
	if len(msg) == 0 {
		return message.Message{}
	}

	if len(msg) == 1 {
		switch msg[0].(type) {
		case string:
			return message.Message{message.Text(msg[0])}
		case []byte:
			return message.Message{message.ImageBytes(msg[0].([]byte))}
		default:
			return message.Message{message.Text("Internal ERROR! Unknown message type")}
		}
	}

	resultMessage := make(message.Message, 0, len(msg))
	for _, node := range msg {
		var correct any
		switch node.(type) {
		case string:
			correct = node
		case []byte:
			correct = message.ImageBytes(node.([]byte))
		default:
			correct = "Internal ERROR! Unknown message type"
		}
		resultMessage = append(resultMessage, message.CustomNode("", 0, correct))
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
					qCtx.SendError(err)
				}
			})
		}
	}

	zero.Run(&zeroCfg)
}
