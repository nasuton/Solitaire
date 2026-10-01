package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// button はヘッダーに並ぶ押しボタン。
type button struct {
	Label   string
	Rect    Rect
	Enabled func() bool
	OnClick func()

	pressed bool // ボタン上で押下が始まった
	hover   bool
}

func (b *button) enabled() bool {
	return b.Enabled == nil || b.Enabled()
}

// update はポインタ状態からクリックを判定する。クリックされたら true。
func (b *button) update(p *pointer) bool {
	inside := b.Rect.Contains(p.X, p.Y)
	b.hover = inside && !p.touching
	if !b.enabled() {
		b.pressed = false
		return false
	}
	if p.JustPressed && inside {
		b.pressed = true
	}
	// 押下と離しが同じ tick でも拾う。
	if p.JustReleased && b.pressed {
		b.pressed = false
		if inside && b.OnClick != nil {
			b.OnClick()
			return true
		}
	}
	if !p.Pressed {
		b.pressed = false
	}
	return false
}

func (b *button) draw(dst *ebiten.Image, scale float64) {
	bg := color.RGBA{0x2b, 0x5a, 0x3e, 0xff}
	fg := color.RGBA{0xf5, 0xf5, 0xf0, 0xff}
	border := color.RGBA{0xc8, 0xd8, 0xc0, 0xff}
	switch {
	case !b.enabled():
		bg = color.RGBA{0x24, 0x44, 0x30, 0xff}
		fg = color.RGBA{0x90, 0xa0, 0x90, 0xff}
		border = color.RGBA{0x60, 0x78, 0x60, 0xff}
	case b.pressed:
		bg = color.RGBA{0x1a, 0x3a, 0x28, 0xff}
	case b.hover:
		bg = color.RGBA{0x3a, 0x74, 0x50, 0xff}
	}
	r := b.Rect
	vector.DrawFilledRect(dst, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), bg, false)
	vector.StrokeRect(dst, float32(r.X)+0.5, float32(r.Y)+0.5, float32(r.W)-1, float32(r.H)-1, 1, border, false)
	drawTextCentered(dst, b.Label, r, scale, fg)
}
