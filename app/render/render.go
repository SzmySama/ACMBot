package render

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"

	"github.com/playwright-community/playwright-go"
	log "github.com/sirupsen/logrus"
)

const ()

var (
	_FULL_TEMPLATE_PATH           string
	codeforcesUserProfileTemplate *template.Template

	_TEMPLATE_PATH                    = "app/templates/"
	codeforcesUserProfileTemplatePath = _TEMPLATE_PATH + "codeforces_profile.html"
)

func init() {
	execPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get exec info: %v", err)
	}
	_FULL_TEMPLATE_PATH = path.Dir(execPath + "/" + _TEMPLATE_PATH)
	log.Infof(_FULL_TEMPLATE_PATH)
	codeforcesUserProfileTemplate, err = template.ParseFiles(codeforcesUserProfileTemplatePath)
	if err != nil {
		log.Fatalf("Failed to load template %s: %v", codeforcesUserProfileTemplatePath, err)
	}

}

func Html(PageOpt *playwright.BrowserNewPageOptions, HTMLOpt *RenderHTMLOptions) ([]byte, error) {
	page, err := GetNewPage(*PageOpt)
	if err != nil {
		return nil, err
	}
	defer page.Close()
	if strings.HasPrefix(HTMLOpt.Path, "file://") {
		page.Goto(HTMLOpt.Path)
	} else {
		page.Goto("file://" + HTMLOpt.Path)
	}
	page.SetContent(HTMLOpt.HTML, playwright.PageSetContentOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	return page.Screenshot(playwright.PageScreenshotOptions{
		// FullPage: &[]bool{true}[0],
		Type: playwright.ScreenshotTypePng,
	})
}

func CodeforcesUserProfile(data CodeforcesUserProfileData) ([]byte, error) {
	var buffer bytes.Buffer
	if err := codeforcesUserProfileTemplate.Execute(&buffer, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %v", err)
	}
	return Html(
		&playwright.BrowserNewPageOptions{
			DeviceScaleFactor: &[]float64{2.0}[0],
			Viewport: &playwright.Size{
				Width:  400,
				Height: 225,
			},
		}, &RenderHTMLOptions{
			Path: _FULL_TEMPLATE_PATH,
			HTML: buffer.String(),
		})
}
