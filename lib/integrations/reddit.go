package integrations

import (
	"fmt"

	"github.com/markbates/goth"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
)

type Reddit struct {
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
	return nil, fmt.Errorf("not yet implemented")
}
