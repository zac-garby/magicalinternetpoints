package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
)

const UserAgent = "magicalinternetpoints/v1 by zacgarby"

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

type ReqOptions struct {
	Method       string
	Data         url.Values
	AuthUsername string
	AuthPassword string
	Headers      map[string]string
}

func getJson(theURL string, target interface{}, opts ...ReqOptions) error {
	var (
		method   = "GET"
		data     = url.Values{}
		username = ""
		password = ""
		headers  = make(map[string]string)
	)

	if len(opts) > 0 {
		opt := opts[0]
		if opt.Method != "" {
			method = opt.Method
		}
		data = opt.Data
		username = opt.AuthUsername
		password = opt.AuthPassword
		if opt.Headers != nil {
			headers = opt.Headers
		}
	}

	req, err := http.NewRequest(method, theURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	if len(username) > 0 {
		req.SetBasicAuth(username, password)
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	// fmt.Printf("%s -> %s\n", theURL, body)

	if err := json.NewDecoder(bytes.NewReader(body)).Decode(target); err != nil {
		all, err2 := io.ReadAll(r.Body)
		if err2 != nil {
			return err2
		}

		if r.StatusCode != 200 {
			return fmt.Errorf("%s (%s): %w", r.Status, string(all), err)
		}

		return err
	}

	return nil
}

func registerIntegration(i Integration) {
	Integrations[i.GetName()] = i
}
