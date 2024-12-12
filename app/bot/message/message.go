package message

import "fmt"

type Type int

const (
	TypeText Type = iota
	TypeImageBytes
	TypeAt
	TypeMixNode
)

type Message []Node

type Node struct {
	MessageType Type
	data        any
}

func (n Node) Text() (string, bool) {
	if n.MessageType != TypeText {
		return "", false
	}
	return n.data.(string), true
}

func (n Node) Text_() string {
	return n.data.(string)
}

func (n Node) ImageBytes() ([]byte, bool) {
	if n.MessageType != TypeImageBytes {
		return nil, false
	}
	return n.data.([]byte), true
}

func (n Node) ImageBytes_() []byte {
	return n.data.([]byte)
}

func (n Node) At() (int64, bool) {
	if n.MessageType != TypeAt {
		return 0, false
	}
	return n.data.(int64), true
}

func (n Node) At_() int64 {
	return n.data.(int64)
}

func (n Node) MixNode() (Node, bool) {
	if n.MessageType != TypeMixNode {
		return Node{}, false
	}
	return n.data.(Node), true
}
func (n Node) MixNode_() Node {
	return n.data.(Node)
}
func Text(text ...any) Node {

	return Node{TypeText, fmt.Sprint(text...)}
}

func ImageBytes(bytes []byte) Node {
	return Node{TypeImageBytes, bytes}
}

func At(at int64) Node {
	return Node{TypeAt, at}
}

func MixNode(node Node) Node {
	return Node{TypeMixNode, node}
}
