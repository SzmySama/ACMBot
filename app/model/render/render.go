package render

import (
	"html/template"
	"os"
	"path"
	"strings"

	"github.com/playwright-community/playwright-go"
	log "github.com/sirupsen/logrus"
)

var (
	_playwright *playwright.Playwright
	_bowers     playwright.Browser

	fullTemplatePath               string
	codeforcesUserProfileTemplate  *template.Template
	codeforcesRatingChangeTemplate *template.Template
)

type Error struct {
	msg string
}

func (e Error) Error() string {
	return "完蛋了，渲染器出错了😰: " + e.msg
}

const (
	templatePath                        = "app/templates/"
	CodeforcesUserProfileTemplatePath   = templatePath + "codeforces_profile.html"
	CodeforcesRatingChangesTemplatePath = templatePath + "codeforces_rating_change.html"
)

type HtmlOptions struct {
	Path string
	HTML string
}

func init() {
	initDriver()
	initTemplates()
}

func initDriver() {
	var err error
	err = playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	})
	if err != nil {
		log.Fatalf("Failed to install playwright: %v", err)
	}
	_playwright, err = playwright.Run()
	if err != nil {
		log.Fatalf("Failed to start playwright: %v", err)
	}

	_bowers, err = _playwright.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		// Headless: &[]bool{false}[0],
	})
	if err != nil {
		log.Fatalf("Failed to launch chromium: %v", err)
	}
	log.Info(_bowers)
}

func initTemplates() {
	execPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get exec info: %v", err)
	}
	fullTemplatePath = path.Dir(execPath + "/" + templatePath)
	log.Infof(fullTemplatePath)

	templateMap := map[**template.Template]string{
		&codeforcesUserProfileTemplate:  CodeforcesUserProfileTemplatePath,
		&codeforcesRatingChangeTemplate: CodeforcesRatingChangesTemplatePath,
	}

	for k, v := range templateMap {
		*k, err = template.ParseFiles(v)
		if err != nil {
			log.Fatalf("Failed to load template %s: %v", v, err)
		}
	}
}

func ShutdownBowers() error {
	return _playwright.Stop()
}

func GetNewPage(opt playwright.BrowserNewPageOptions) (playwright.Page, error) {
	return _bowers.NewPage(opt)
}

func Html(PageOpt *playwright.BrowserNewPageOptions, HTMLOpt *HtmlOptions) ([]byte, error) {
	page, err := GetNewPage(*PageOpt)
	if err != nil {
		return nil, Error{msg: err.Error()}
	}
	defer page.Close()

	if strings.HasPrefix(HTMLOpt.Path, "file://") {
		HTMLOpt.Path = "file://" + HTMLOpt.Path
	}
	if _, err = page.Goto("file://" + HTMLOpt.Path); err != nil {
		return nil, Error{msg: err.Error()}
	}
	if err = page.SetContent(HTMLOpt.HTML, playwright.PageSetContentOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		return nil, Error{msg: err.Error()}
	}
	data, err := page.Screenshot(playwright.PageScreenshotOptions{
		// FullPage: &[]bool{true}[0],
		Type: playwright.ScreenshotTypePng,
	})
	if err != nil {
		return nil, Error{msg: err.Error()}
	}
	return data, nil
}
