package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/nasuton/Solitaire/internal/uifont"
)

// drawText は (x, y) を左上として文字列を描く。
func drawText(dst *ebiten.Image, s string, x, y, scale float64, clr color.Color) {
	uifont.Draw(dst, s, x, y, scale, clr)
}

// drawTextCentered は矩形の中央に文字列を描く。
func drawTextCentered(dst *ebiten.Image, s string, r Rect, scale float64, clr color.Color) {
	w, h := textSize(s, scale)
	drawText(dst, s, r.X+(r.W-w)/2, r.Y+(r.H-h)/2, scale, clr)
}

// textSize は拡大後の文字列サイズを返す。
func textSize(s string, scale float64) (w, h float64) {
	return uifont.Measure(s, scale)
}
