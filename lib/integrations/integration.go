package integrations

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
)

var Integrations map[string]Integration = make(map[string]Integration)

type Integration interface {
	// returns the site name as per the 'title' database field
	GetName() string

	// get the authentication provider for this site integration
	GetAuthProvider() AuthProvider

	// gets the user profile URL for a username
	GetProfileURL(username string) string

	// makes API calls to calculate the raw point totals of a user
	GetRawPoints(*common.Account) (map[string]int, error)
}

type AuthProvider interface {
	// begins authentication for some user for this site. probably
	// redirects to some new page where either OAuth happens or
	// they enter some authenticating information.
	BeginAuthentication(user *common.User, c *fiber.Ctx) error
}

func getJson(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		if r.StatusCode != 200 {
			return fmt.Errorf("non-200 error: %w", err)
		}

		return err
	}

	return nil
}

func registerIntegration(i Integration) {
	Integrations[i.GetName()] = i
}
