// Package assets は埋め込み画像をデコードして ebiten.Image として提供する。
package assets

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png" // PNG デコーダ登録

	"github.com/hajimehoshi/ebiten/v2"

	root "github.com/nasuton/Solitaire/assets"
	"github.com/nasuton/Solitaire/internal/cardfiles"
	"github.com/nasuton/Solitaire/internal/klondike"
)

// Cards はカード画像一式。
type Cards struct {
	faces [52]*ebiten.Image
	back  *ebiten.Image
	// Width / Height は画像のピクセルサイズ（全カード共通）。
	Width, Height int
}

// Load は埋め込み画像をすべてデコードする。
func Load() (*Cards, error) {
	c := &Cards{}
	for _, s := range klondike.Suits {
		for r := klondike.Ace; r <= klondike.King; r++ {
			img, err := decode(cardfiles.Path(s, r))
			if err != nil {
				return nil, err
			}
			c.faces[klondike.Card{Suit: s, Rank: r}.ID()] = img
		}
	}
	back, err := decode(cardfiles.BackPath())
	if err != nil {
		return nil, err
	}
	c.back = back
	b := back.Bounds()
	c.Width, c.Height = b.Dx(), b.Dy()
	return c, nil
}

// Face は card の表面画像を返す。
func (c *Cards) Face(card klondike.Card) *ebiten.Image {
	return c.faces[card.ID()]
}

// Back は裏面画像を返す。
func (c *Cards) Back() *ebiten.Image {
	return c.back
}

// Image は card の向きに応じて表面または裏面を返す。
func (c *Cards) Image(card klondike.Card) *ebiten.Image {
	if card.FaceUp {
		return c.Face(card)
	}
	return c.back
}

func decode(path string) (*ebiten.Image, error) {
	data, err := root.FS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("assets: read %s: %w", path, err)
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("assets: decode %s: %w", path, err)
	}
	return ebiten.NewImageFromImage(img), nil
}
