package errs

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

type InternalError struct {
	frames []runtime.Frame
	text   string
}

func (e InternalError) Error() string {
	var stackTrace strings.Builder
	for _, frame := range e.frames {
		stackTrace.WriteString(fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function))
	}
	return fmt.Sprintf("INTERNAL ERROR! %s\nstack:\n%s", e.text, stackTrace.String())
}

func NewInternalError(message string) InternalError {
	return InternalError{
		text:   message,
		frames: GetCallStack(),
	}
}

func GetCallStack() []runtime.Frame {
	var buf []uintptr
	n := runtime.Callers(3, buf[:0]) // 跳过 GetCallStack 和它的调用者
	buf = make([]uintptr, n)
	n = runtime.Callers(3, buf)

	var frames []runtime.Frame
	for _, pc := range buf {
		frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
		frames = append(frames, frame)
	}
	return frames
}

var (
	ErrNoRatingChanges       = errors.New("没有找到任何Rating变化记录哦，可能分还没出来，总不可能你没打过比赛吧... ")
	ErrHandleNotFound        = errors.New("没有叫这个名字的用户哦，是不是打错了？")
	ErrNoHandle              = errors.New("没有听到要查询谁哦")
	ErrGroupOnly             = errors.New("该功能必须要在群内使用哦")
	ErrImDedicated           = errors.New("需要且仅一个参数哦，不要发无效信息啦~")
	ErrBadPlatform           = errors.New("暂不支持该平台哦")
	ErrOrganizationUnmatched = errors.New("绑定前请前往https://codeforces.com/settings/social,将Organization设置为`ACMBot`")
	ErrHandleHasBindByOthers = errors.New("该codeforces账号已被他人绑定了哦")
)
