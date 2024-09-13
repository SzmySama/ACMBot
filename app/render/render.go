package render

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"

	"github.com/SzmySama/ACMBot/app/types"
	"github.com/playwright-community/playwright-go"
	log "github.com/sirupsen/logrus"
)

const ()

var (
	_FULL_TEMPLATE_PATH            string
	codeforcesUserProfileTemplate  *template.Template
	codeforcesRatingChangeTemplate *template.Template
)

const (
	_TEMPLATE_PATH                          = "app/templates/"
	CODEFORCES_USER_PROFILE_TEMPLATE_PATH   = _TEMPLATE_PATH + "codeforces_profile.html"
	CODEFORCES_RATING_CHANGES_TEMPLATE_PATH = _TEMPLATE_PATH + "codeforces_rating_change.html"
)

func init() {
	execPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get exec info: %v", err)
	}
	_FULL_TEMPLATE_PATH = path.Dir(execPath + "/" + _TEMPLATE_PATH)
	log.Infof(_FULL_TEMPLATE_PATH)

	template_map := map[**template.Template]string{
		&codeforcesUserProfileTemplate:  CODEFORCES_USER_PROFILE_TEMPLATE_PATH,
		&codeforcesRatingChangeTemplate: CODEFORCES_RATING_CHANGES_TEMPLATE_PATH,
	}

	for k, v := range template_map {
		*k, err = template.ParseFiles(v)
		if err != nil {
			log.Fatalf("Failed to load template %s: %v", v, err)
		}
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

func CodeforcesUserProfile(user types.User) ([]byte, error) {
	var buffer bytes.Buffer
	if err := codeforcesUserProfileTemplate.Execute(&buffer, CodeforcesUserProfileData{
		User:  user,
		Level: ConvertRatingToLevel(user.Rating),
	}); err != nil {
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
		},
	)
}

func CodeforcesRatingChanges(ratingChanges []types.RatingChange, handle string) ([]byte, error) {
	var buffer bytes.Buffer
	if err := codeforcesRatingChangeTemplate.Execute(&buffer, CodeforcesRatingChangesData{
		RatingChangesMetaData: ratingChanges,
		Handle:                handle,
	}); err != nil {
		return nil, fmt.Errorf("failed to execute template: %v", err)
	}
	return Html(
		&playwright.BrowserNewPageOptions{
			DeviceScaleFactor: &[]float64{2.0}[0],
			Viewport: &playwright.Size{
				Width:  1000,
				Height: 500,
			},
		}, &RenderHTMLOptions{
			Path: _FULL_TEMPLATE_PATH,
			HTML: buffer.String(),
		},
	)
}
