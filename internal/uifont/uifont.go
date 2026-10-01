// Package uifont は tools/fontgen が生成したグリフアトラスで UI 文字列を描画する。
// 元フォントは github.com/hajimehoshi/bitmapfont/v4 の FaceEA（12px、日本語対応）。
package uifont

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	_ "image/png" // PNG デコーダ登録

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed atlas.png
var atlasPNG []byte

// glyph はアトラス内の位置（index）と送り幅（adv）。
type glyph struct {
	index int
	adv   int
}

var atlas *ebiten.Image

func ensureAtlas() *ebiten.Image {
	if atlas != nil {
		return atlas
	}
	img, _, err := image.Decode(bytes.NewReader(atlasPNG))
	if err != nil {
		panic("uifont: decode atlas: " + err.Error())
	}
	atlas = ebiten.NewImageFromImage(img)
	return atlas
}

// LineHeight は拡大前の行高（px）。
func LineHeight() float64 { return cellH }

// Measure は拡大率 scale での文字列の幅と高さを返す。
func Measure(s string, scale float64) (w, h float64) {
	adv := 0
	for _, r := range s {
		adv += lookup(r).adv
	}
	return float64(adv) * scale, cellH * scale
}

// Has は文字 r がアトラスに含まれるかを返す。
func Has(r rune) bool {
	_, ok := glyphs[r]
	return ok
}

func lookup(r rune) glyph {
	if g, ok := glyphs[r]; ok {
		return g
	}
	if g, ok := glyphs['\uFFFD']; ok {
		return g
	}
	return glyphs['?']
}

// Draw は (x, y) を左上として文字列を描画する。
func Draw(dst *ebiten.Image, s string, x, y, scale float64, clr color.Color) {
	if s == "" {
		return
	}
	a := ensureAtlas()
	penX := x
	for _, r := range s {
		g := lookup(r)
		cx := (g.index % columns) * cellW
		cy := (g.index / columns) * cellH
		sub := a.SubImage(image.Rect(cx, cy, cx+g.adv, cy+cellH)).(*ebiten.Image)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(penX, y)
		op.ColorScale.ScaleWithColor(clr)
		op.Filter = ebiten.FilterNearest
		dst.DrawImage(sub, op)
		penX += float64(g.adv) * scale
	}
}
