package providers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/zac-garby/magicalinternetpoints/lib/common"
)

type OAuthProvider struct {
	ProviderName string
}

func (o *OAuthProvider) BeginAuthentication(user *common.User, c *fiber.Ctx) error {
	return c.Redirect(fmt.Sprintf("/auth/begin-oauth/%s", o.ProviderName))
}
