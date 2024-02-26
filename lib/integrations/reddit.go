package integrations

import (
	"fmt"
	"net/url"
	"os"

	"github.com/zac-garby/magicalinternetpoints/lib/common"
	"github.com/zac-garby/magicalinternetpoints/lib/integrations/providers"
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

type redditAccess struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func init() {
	registerIntegration(&Reddit{})
}

func (r *Reddit) GetName() string {
	return "Reddit"
}

func (r *Reddit) GetAuthProvider() AuthProvider {
	return &providers.OAuthProvider{
		ProviderName: "reddit",
	}
}

func (r *Reddit) GetProfileURL(username string) string {
	return fmt.Sprintf("https://reddit.com/u/%s", username)
}

func (r *Reddit) GetRawPoints(account *common.Account) (map[string]int, error) {
	auth, err := r.getAuth()
	if err != nil {
		return nil, fmt.Errorf("could not authenticate with reddit: %w", err)
	}

	url := fmt.Sprintf("https://oauth.reddit.com/user/%s/about", account.Username)

	resp := redditResponse{}
	if err := getJson(url, &resp, ReqOptions{
		Headers: map[string]string{
			"Authorization": fmt.Sprintf("bearer %s", auth.AccessToken),
			"User-agent":    UserAgent,
		},
	}); err != nil {
		return nil, fmt.Errorf("error getting json for reddit user %s: %w", account.Username, err)
	}

	return map[string]int{
		"link karma":    resp.Data.LinkKarma,
		"comment karma": resp.Data.CommentKarma,
	}, nil
}

func (r *Reddit) getAuth() (*redditAccess, error) {
	u := "https://www.reddit.com/api/v1/access_token"

	v := url.Values{}
	v.Set("grant_type", "client_credentials")

	var data redditAccess
	if err := getJson(u, &data, ReqOptions{
		Method:       "POST",
		Data:         v,
		AuthUsername: os.Getenv("REDDIT_TOKEN"),
		AuthPassword: os.Getenv("REDDIT_SECRET"),
		Headers: map[string]string{
			"User-agent": UserAgent,
		},
	}); err != nil {
		return nil, nil
	}

	return &data, nil
}
