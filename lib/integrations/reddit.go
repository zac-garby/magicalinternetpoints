package integrations

import (
	"context"
	"fmt"

	"github.com/zac-garby/magicalinternetpoints/lib/common"
	"github.com/zac-garby/magicalinternetpoints/lib/integrations/providers"

	"github.com/vartanbeno/go-reddit/v2/reddit"
)

type Reddit struct {
	client *reddit.Client
}

func init() {
	client, err := reddit.NewReadonlyClient()
	if err != nil {
		panic(err)
	}

	registerIntegration(&Reddit{
		client: client,
	})
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
	user, _, err := r.client.User.Get(context.Background(), account.Username)
	if err != nil {
		return nil, err
	}

	return map[string]int{
		"link karma":    user.PostKarma,
		"comment karma": user.CommentKarma,
	}, nil
}
