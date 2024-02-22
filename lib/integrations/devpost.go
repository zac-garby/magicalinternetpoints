package integrations

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
	"github.com/zac-garby/magicalinternetpoints/lib/integrations/providers"
)

var (
	DEVPOST_QUERIES = map[string]string{
		"hacks":     "a[href='/%s'] > div.totals > span",
		"followers": "a[href='/%s/followers'] > div.totals > span",
		"wins":      "img.winner",
		"likes":     "span.like-count",
	}
)

type Devpost struct {
}

func init() {
	registerIntegration(&Devpost{})
}

func (d *Devpost) GetName() string {
	return "Devpost"
}

func (d *Devpost) GetAuthProvider() AuthProvider {
	return &providers.BioAuthProvider{
		ProfileURL:  "https://devpost.com/%s",
		BioSelector: "p#portfolio-user-bio",
	}
}

func (d *Devpost) GetProfileURL(username string) string {
	return fmt.Sprintf("https://devpost.com/%s", username)
}

func (d *Devpost) GetRawPoints(account *common.Account) (map[string]int, error) {
	var (
		col = colly.NewCollector()
		raw = map[string]int{
			"wins":      0,
			"hacks":     0,
			"followers": 0,
			"likes":     0,
		}
		finalErr error = nil
	)

	profileURL := d.GetProfileURL(account.Username)

	for k, sel := range DEVPOST_QUERIES {
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

			n, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)

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
