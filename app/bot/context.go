package bot

import "github.com/YourSuzumiya/ACMBot/app/bot/message"

type Context struct {
	Invoker
	Platform Platform

	StepValue any
}

type Invoker interface {
	Send(message message.Message)
	SendError(err error)
	Params() []string
	GetCallerInfo() CallerInfo
}

type CallerInfo struct {
	NickName string
	ID       int64
	Group    GroupInfo
}

type GroupInfo struct {
	ID          int64
	Name        string
	MemberCount int64
}

type Platform int

const (
	PlatformQQ Platform = iota
)
