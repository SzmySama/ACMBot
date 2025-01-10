package message

import "github.com/YourSuzumiya/ACMBot/app/model"

type Message interface {
	__msg()
}

type Text string

func (t Text) __msg() {}

type Image []byte

func (i Image) __msg() {}

type Races []model.Race

func (r Races) __msg() {}
