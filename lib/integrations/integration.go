package integrations

import (
	"encoding/json"
	"net/http"

	"github.com/markbates/goth"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
)

var Integrations map[string]Integration = make(map[string]Integration)

type Integration interface {
	// returns the site name as per the 'title' database field
	GetName() string

	// gets the OAuth provider name (referring to goth providers)
	GetOAuthProvider() string

	// gets the user profile URL for an OAuth user
	GetProfileURL(user *goth.User) string

	// makes API calls to calculate the raw point totals of a user
	GetRawPoints(*common.Account) (map[string]int, error)
}

func getJson(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func registerIntegration(i Integration) {
	Integrations[i.GetName()] = i
}
