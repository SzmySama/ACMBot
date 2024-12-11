package bot

type Context struct {
	Invoker
	ProtoType ProtoType

	StepValue any
}

type Invoker interface {
	Send(message Message)
	SendError(err error)
	Params() Message
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

type MessageType int

type Message []MessageNode

type MessageNode any
