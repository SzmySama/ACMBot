package qq

import (
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app"
	"github.com/YourSuzumiya/ACMBot/app/bot"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type qqContext struct {
	zCtx *zero.Ctx
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
	logrus.Infof(argStr)
	return bot.Message{}
}

func msgToZeroMsg(msg bot.Message) message.Message {
	if len(msg) == 0 {
		return message.Message{message.Text("刚才好像有人叫我来着...可是萝卜子已经忘记自己要说什么了")}
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

func init() {
	cfg := app.GetConfig().Bot
	zeroCfg = zero.Config{
		NickName:      cfg.NickName,
		CommandPrefix: cfg.CommandPrefix,
		SuperUsers:    cfg.SuperUsers,
		Driver:        []zero.Driver{},
	}

	for _, cfg := range cfg.WS {
		zeroCfg.Driver = append(zeroCfg.Driver, driver.NewWebSocketClient(
			fmt.Sprintf("ws://%s:%d", cfg.Host, cfg.Port),
			cfg.Token))
	}

	for command, handler := range bot.CommandMap {
		zero.OnCommand(command).Handle(func(ctx *zero.Ctx) {
			handler(&qqContext{zCtx: ctx})
		})
	}
}
