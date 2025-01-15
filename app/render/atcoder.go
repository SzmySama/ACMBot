package render

import (
	"bytes"
	"fmt"

	"github.com/playwright-community/playwright-go"
)

type AtcoderSolvedData struct {
	Range   string
	Count   uint
	Percent float64
}

type AtcoderUserProfile struct {
	Avatar    string
	Handle    string
	MaxRating uint
	Rating    uint
	Level     string
	Solved    uint
    PromotionMessage string
    Time string

	SolvedData []AtcoderSolvedData
}

func (u *AtcoderUserProfile) ToImage() ([]byte, error) {
	var buffer bytes.Buffer
	if err := atcoderUserProfileTemplate.Execute(&buffer, u); err != nil {
		return nil, Error{fmt.Sprintf("failed to execute template: %v", err)}
	}

	return Html(
		&playwright.BrowserNewPageOptions{
			DeviceScaleFactor: &[]float64{2.0}[0],
			Viewport: &playwright.Size{
				Width:  300,
				Height: 400,
			},
		}, &HtmlOptions{
			Path: fullTemplatePath,
			HTML: buffer.String(),
		},
	)
}
