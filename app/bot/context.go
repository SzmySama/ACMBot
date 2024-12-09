package bot

/*
Context

	Send发送消息
	如果数组长度大于1，则发送为合并转发消息

	SendError
	如果bot出错了是不是应该告诉开发者一声?在这里写告诉他的逻辑吧

	Params获取用户提供的参数
	忽略@
	字符串以空格分割
*/
type Context interface {
	Send(message Message)
	SendError(err error)
	Params() Message
}

type MessageType int

type Message []MessageNode

type MessageNode any

type Handler func(ctx Context)
