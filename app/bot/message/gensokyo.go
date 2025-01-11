package message

import (
	"encoding/base64"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func MarkDown(raw map[string]any) message.MessageSegment {
	/*
		MarkDown(map[string]any{
					"markdown": map[string]any{
						"content": "# Test Title\n - l1\n - l2",
					},
				}),
	*/
	res, _ := json.Marshal(raw)
	logrus.Info(string(res))
	return message.MessageSegment{
		Type: "markdown",
		Data: map[string]string{
			"data": `{"data":"base64://` + base64.StdEncoding.EncodeToString(res) + `"}`,
		},
	}
}
