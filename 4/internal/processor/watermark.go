package processor

import (
	"image"
	"image/color"
	"log"
	"os"

	"github.com/disintegration/imaging"
	"github.com/golang/freetype"
)

func AddTextWatermark(img image.Image, text, fontPath string) *image.NRGBA {
	rgba := imaging.Clone(img)

	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		log.Println("font not found:", err)
		return rgba
	}

	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println("font parse error:", err)
		return rgba
	}

	w := rgba.Bounds().Dx()
	h := rgba.Bounds().Dy()

	ctx := freetype.NewContext()
	ctx.SetFont(f)
	ctx.SetFontSize(float64(w) / 15)
	ctx.SetDst(rgba)
	ctx.SetClip(rgba.Bounds())
	ctx.SetSrc(image.NewUniform(color.RGBA{255, 255, 255, 200}))

	pt := freetype.Pt(20, h-40)
	_, err = ctx.DrawString(text, pt)
	if err != nil {
		log.Println("draw error:", err)
	}

	return rgba
}
