package errs

import (
	"errors"
	"fmt"
	"github.com/go-stack/stack"
	"runtime"
	"strings"
)

type InternalError struct {
	stack stack.CallStack
	text  string
}

func NewInternalError(message string) error {
	stackTrace := GetCallStack()
	return errors.New(fmt.Sprintf("%s\nStack Trace:\n%s", message, stackTrace))
}

func GetCallStack() string {
	var buf []uintptr
	n := runtime.Callers(3, buf[:0]) // 跳过 GetCallStack 和它的调用者
	buf = make([]uintptr, n)
	n = runtime.Callers(3, buf)

	var frames []runtime.Frame
	for _, pc := range buf {
		frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
		frames = append(frames, frame)
	}

	var stackTrace strings.Builder
	for _, frame := range frames {
		stackTrace.WriteString(fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function))
	}
	return stackTrace.String()
}

var (
	ErrNoRatingChanges = errors.New("没有找到任何Rating变化记录哦，可能分还没出来，总不可能你没打过比赛吧... ")
	ErrHandleNotFound  = errors.New("没有叫这个名字的用户哦，是不是打错了？")
	ErrNoHandle        = errors.New("没有听到要查询谁哦")
	ErrBadParam        = errors.New("你好像发了什么不得了的东西...暂时无法处理哦")

	ErrBadBranch = errors.New("internal ERROR! unexpected branch")

	ErrUninit = errors.New("internal error! Datastructures didn't init")
)
