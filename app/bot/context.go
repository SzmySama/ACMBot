package bot

import "github.com/YourSuzumiya/ACMBot/app/bot/message"

type Context struct {
	ApiCaller
	Platform Platform

	StepValue any
}

type ApiCaller interface {
	Send(message message.Message)
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
