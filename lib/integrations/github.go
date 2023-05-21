package integrations

import (
	"fmt"

	"github.com/markbates/goth"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
)

type GitHub struct {
}

type jsonGitHubRepo struct {
	Name       string              `json:"name"`
	Owner      jsonGitHubRepoOwner `json:"owner"`
	Stargazers int                 `json:"stargazers_count"`
}

type jsonGitHubRepoOwner struct {
	Login string `json:"login"`
}

type jsonGitHubUser struct {
	Followers int `json:"followers"`
}

func init() {
	registerIntegration(&GitHub{})
}

func (g *GitHub) GetName() string {
	return "GitHub"
}

func (g *GitHub) GetOAuthProvider() string {
	return "github"
}

func (g *GitHub) GetProfileURL(user *goth.User) string {
	return fmt.Sprintf("https://github.com/%s", user.NickName)
}

func (g *GitHub) GetRawPoints(account *common.Account) (map[string]int, error) {
	stars, err := g.getStars(account)
	if err != nil {
		return nil, fmt.Errorf("error getting github stars: %w", err)
	}

	followers, err := g.getFollowers(account)
	if err != nil {
		return nil, fmt.Errorf("error getting github followers: %w", err)
	}

	return map[string]int{
		"stars":     stars,
		"followers": followers,
	}, nil
}

func (g *GitHub) getStars(account *common.Account) (int, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=100&sort=updated", account.Username)

	resp := make([]jsonGitHubRepo, 100)
	if err := getJson(url, &resp); err != nil {
		return -1, fmt.Errorf("error getting json for repos: %w", err)
	}

	stars := 0
	for _, repo := range resp {
		if repo.Owner.Login == account.Username {
			stars += repo.Stargazers
		}
	}

	return stars, nil
}

func (g *GitHub) getFollowers(account *common.Account) (int, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s", account.Username)

	resp := jsonGitHubUser{}
	if err := getJson(url, &resp); err != nil {
		return -1, fmt.Errorf("error getting json for account: %w", err)
	}

	return resp.Followers, nil
}
