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

var (
	fullTemplatePath               string
	codeforcesUserProfileTemplate  *template.Template
	codeforcesRatingChangeTemplate *template.Template
)

const (
	templatePath                        = "app/templates/"
	CodeforcesUserProfileTemplatePath   = templatePath + "codeforces_profile.html"
	CodeforcesRatingChangesTemplatePath = templatePath + "codeforces_rating_change.html"
)

func init() {
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

func Html(PageOpt *playwright.BrowserNewPageOptions, HTMLOpt *HtmlOptions) ([]byte, error) {
	page, err := GetNewPage(*PageOpt)
	if err != nil {
		return nil, err
	}
	defer func(page playwright.Page, options ...playwright.PageCloseOptions) {
		err := page.Close(options...)
		if err != nil {
			log.Errorf("Failed to close page: %v", err)
		}
	}(page)
	if strings.HasPrefix(HTMLOpt.Path, "file://") {
		_, err := page.Goto(HTMLOpt.Path)
		if err != nil {
			return nil, err
		}
	} else {
		_, err := page.Goto("file://" + HTMLOpt.Path)
		if err != nil {
			return nil, err
		}
	}
	err = page.SetContent(HTMLOpt.HTML, playwright.PageSetContentOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return nil, err
	}
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
		}, &HtmlOptions{
			Path: fullTemplatePath,
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
		}, &HtmlOptions{
			Path: fullTemplatePath,
			HTML: buffer.String(),
		},
	)
}
