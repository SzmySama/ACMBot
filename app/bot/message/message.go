package message

import (
	"github.com/YourSuzumiya/ACMBot/app/model"
	zMsg "github.com/wdvxdr1123/ZeroBot/message"
)

type Message interface {
	ToZeroMessage() zMsg.Message
}

type Text string

func (t Text) ToZeroMessage() zMsg.Message {
	return zMsg.Message{zMsg.Text(t)}
}

type Image []byte

func (i Image) ToZeroMessage() zMsg.Message {
	return zMsg.Message{zMsg.ImageBytes(i)}
}

type Races []model.Race

func (r Races) ToZeroMessage() zMsg.Message {
	var result string

	for _, race := range r[:min(8, len(r))] {
		result += "\n" + race.NoUrlString()
	}
	return zMsg.Message{zMsg.Text(result)}
}
