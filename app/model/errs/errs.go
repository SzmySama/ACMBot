package errs

import "errors"

var (
	ErrNoRatingChanges = errors.New("没有找到任何Rating变化记录哦，可能分还没出来，总不可能你没打过比赛吧... ")
	ErrHandleNotFound  = errors.New("没有叫这个名字的用户哦，是不是打错了？")

	ErrUninit = errors.New("internal error! Datastructures didn't init")
)
