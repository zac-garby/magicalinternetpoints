package integrations

import (
	"fmt"

	"github.com/markbates/goth"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
)

type Reddit struct {
}

type redditResponse struct {
	Data redditUserData `json:"data"`
}

type redditUserData struct {
	LinkKarma    int `json:"link_karma"`
	CommentKarma int `json:"comment_karma"`
}

func init() {
	registerIntegration(&Reddit{})
}

func (r *Reddit) GetName() string {
	return "Reddit"
}

func (r *Reddit) GetOAuthProvider() string {
	return "reddit"
}

func (r *Reddit) GetProfileURL(user *goth.User) string {
	return fmt.Sprintf("https://reddit.com/u/%s", user.NickName)
}

func (r *Reddit) GetRawPoints(account *common.Account) (map[string]int, error) {
	url := fmt.Sprintf("https://reddit.com/user/%s/about.json", account.Username)

	resp := redditResponse{}
	if err := getJson(url, &resp); err != nil {
		return nil, fmt.Errorf("error getting json for reddit user %s: %w", account.Username, err)
	}

	return map[string]int{
		"link karma":    resp.Data.LinkKarma,
		"comment karma": resp.Data.CommentKarma,
	}, nil
}
