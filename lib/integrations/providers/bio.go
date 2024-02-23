package providers

import (
	"fmt"
	"html/template"

	"github.com/gofiber/fiber/v2"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
)

const BIO_AUTH_STRING = "[mip_verify:%s]"

type BioAuthProvider struct {
	// the name of the "bio". in other words, while we assume that in
	// most cases it's the user's biography where the text will be, sometimes
	// that's not possible and some other term must be used
	BioLanguage string

	// the term by which we refer to the "username", since this isn't always
	// the correct word for all websites.
	UsernameLanguage string

	// extra instructions to be displayed on the "enter username" page
	ExtraUsernameInstructions template.HTML

	// extra instructions to be displayed on the page which shows the verify
	// text to enter
	ExtraVerifyInstructions template.HTML

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
