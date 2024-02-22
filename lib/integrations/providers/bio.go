package providers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
)

const BIO_AUTH_STRING = "[mip_verify:%s]"

type BioAuthProvider struct {
	// the url of the user's profile, with %s substituted for
	// their username.
	ProfileURL string

	// a CSS selector to locate the element containing their
	// bio. all child elements' text will be concatenated together
	BioSelector string
}

func (b *BioAuthProvider) BeginAuthentication(user *common.User, c *fiber.Ctx) error {
	return c.Redirect(fmt.Sprintf("/auth/bio/%s", c.Params("site_title")))
}
