package integrations

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
	"github.com/zac-garby/magicalinternetpoints/lib/integrations/providers"
)

var (
	GOODREADS_QUERIES = map[string]string{
		"read":  "#shelves a[href$='?shelf=read']",
		"rated": "div.leftAlignedProfilePicture a[href$='?sort=rating&view=reviews']",
	}
)

type Goodreads struct {
}

func init() {
	registerIntegration(&Goodreads{})
}

func (d *Goodreads) GetName() string {
	return "Goodreads"
}

func (d *Goodreads) GetAuthProvider() AuthProvider {
	return &providers.BioAuthProvider{
		ProfileURL:                "https://www.goodreads.com/%s",
		BioSelector:               "div.userInfoBoxContent a[rel^=me]",
		BioLanguage:               "'My Web Site' url",
		UsernameLanguage:          "profile url",
		ExtraUsernameInstructions: template.HTML("your profile url can be found on the <a href=\"https://www.goodreads.com/user/edit?tab=profile\">settings</a> page, under \"Profile URL\"."),
		ExtraVerifyInstructions:   template.HTML("your 'My Web Site' url can be found on the <a href=\"https://www.goodreads.com/user/edit?tab=profile\">settings</a> page. if you don't want to overwrite this, even for a moment, you can post-append the string after a hash '#' so the current URL remains."),
	}
}

func (d *Goodreads) GetProfileURL(username string) string {
	return fmt.Sprintf("https://www.goodreads.com/%s", username)
}

func (d *Goodreads) GetRawPoints(account *common.Account) (map[string]int, error) {
	var (
		col = colly.NewCollector()
		raw = map[string]int{
			"read":  0,
			"rated": 0,
		}
		finalErr error = nil
	)

	profileURL := d.GetProfileURL(account.Username)

	for k, sel := range GOODREADS_QUERIES {
		key := strings.Clone(k)

		selector := sel
		if strings.Contains(sel, "%s") {
			selector = fmt.Sprintf(sel, account.Username)
		}

		col.OnHTML(selector, func(e *colly.HTMLElement) {
			text := e.Text
			if text == "" {
				text = e.ChildText("*")
			}

			n, err := strconv.ParseInt(strip(text), 10, 64)

			if err != nil {
				raw[key]++
			} else {
				raw[key] += int(n)
			}
		})
	}

	col.Visit(profileURL)
	col.Wait()

	if finalErr != nil {
		return nil, finalErr
	}

	return raw, nil
}

func strip(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if '0' <= b && b <= '9' {
			result.WriteByte(b)
		}
	}
	return result.String()
}
