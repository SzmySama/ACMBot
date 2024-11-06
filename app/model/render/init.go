package render

import (
	"github.com/playwright-community/playwright-go"
	log "github.com/sirupsen/logrus"
)

var (
	_playwright *playwright.Playwright
	_bowers     playwright.Browser
)

func init() {
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

func ShutdownBowers() error {
	return _playwright.Stop()
}

func GetNewPage(opt playwright.BrowserNewPageOptions) (playwright.Page, error) {
	return _bowers.NewPage(opt)
}
