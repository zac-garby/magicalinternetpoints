package backend

import (
	"fmt"

	"github.com/fogleman/gg"
	"github.com/gofiber/fiber/v2"
)

const (
	width  = 320
	height = 36
)

func (b *Backend) GetBadge(c *fiber.Ctx) error {
	id, err := b.GetUserID(c.Params("username"))
	if err != nil {
		return err
	}

	total, _, err := b.GetRawPoints(id)
	if err != nil {
		return err
	}

	dc := gg.NewContext(width, height)

	if err := b.drawBadge(dc, total); err != nil {
		return err
	}

	c.Response().Header.SetContentType("image/png")
	return dc.EncodePNG(c.Response().BodyWriter())
}

func (b *Backend) drawBadge(dc *gg.Context, points int) error {
	dc.SetRGB255(235, 218, 196)
	dc.DrawRoundedRectangle(0, 0, width, height, 12)
	dc.Fill()

	dc.SetRGB255(187, 173, 156)
	dc.SetLineWidth(4)
	dc.DrawRoundedRectangle(2, 2, width-4, height-4, 8)
	dc.Stroke()

	if err := dc.LoadFontFace("./static/fonts/inconsolata-condensed.ttf", 20); err != nil {
		return err
	}

	dc.SetRGB255(60, 10, 43)
	dc.DrawStringAnchored(
		fmt.Sprintf("i have %d magicalinternetpoints :)", points),
		12, height/2, 0, 0.4,
	)

	return nil
}
