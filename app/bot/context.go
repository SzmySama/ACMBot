package bot

import "github.com/YourSuzumiya/ACMBot/app/bot/message"

type Context struct {
	Invoker
	ProtoType ProtoType

	StepValue any
}

type Invoker interface {
	Send(message message.Message)
	SendError(err error)
	Params() message.Message
	GetSender() SenderInfo
}

type SenderInfo struct {
	ID       int64
	GroupID  int64
	NickName string
}

type ProtoType int

const (
	ProtoTypeQQ ProtoType = iota
)
