// Package ui は Ebitengine を使ったソリティアの描画・入力・アニメーションを担当する。
package ui

import (
	"math"

	"github.com/nasuton/Solitaire/internal/klondike"
)

// CardAspect はカード画像の高さ / 幅（元画像 712x1008）。
const CardAspect = 1008.0 / 712.0

// Rect は論理座標上の矩形。
type Rect struct {
	X, Y, W, H float64
}

// Contains は点 (x, y) が矩形内かを返す。
func (r Rect) Contains(x, y float64) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

// OverlapArea は 2 つの矩形の重なり面積を返す。
func (r Rect) OverlapArea(o Rect) float64 {
	w := math.Min(r.X+r.W, o.X+o.W) - math.Max(r.X, o.X)
	h := math.Min(r.Y+r.H, o.Y+o.H) - math.Max(r.Y, o.Y)
	if w <= 0 || h <= 0 {
		return 0
	}
	return w * h
}

// Layout は論理画面サイズから求めた各置き場の配置。
type Layout struct {
	W, H         float64
	Gap          float64
	HeaderH      float64
	TwoRowHeader bool
	CardW, CardH float64

	Stock       Rect
	Waste       Rect
	Foundations [4]Rect
	Tableau     [7]Rect // 各列の先頭カード位置

	// 場札で重ねる際の既定オフセット。
	UpOffset   float64
	DownOffset float64
	// 場札が使える下端。
	Bottom float64

	// テキストの拡大率（bitmapfont 12px 基準）。
	TextScale float64
}

// NewLayout は論理サイズ w x h に対する配置を計算する。
func NewLayout(w, h float64) Layout {
	l := Layout{W: w, H: h}
	l.Gap = math.Max(8, w*0.012)
	l.TextScale = 2
	l.TwoRowHeader = w < 1000
	line := 13*l.TextScale + 8
	if l.TwoRowHeader {
		l.HeaderH = line*2 + 20
	} else {
		l.HeaderH = line + 20
	}

	widthLimited := (w - 8*l.Gap) / 7
	avail := h - l.HeaderH - 3*l.Gap
	heightLimited := avail / 4.2 / CardAspect
	l.CardW = math.Floor(math.Min(widthLimited, heightLimited))
	if l.CardW < 24 {
		l.CardW = 24
	}
	l.CardH = math.Floor(l.CardW * CardAspect)
	l.UpOffset = math.Floor(l.CardH * 0.28)
	l.DownOffset = math.Floor(l.CardH * 0.12)

	total := 7*l.CardW + 6*l.Gap
	x0 := math.Floor((w - total) / 2)
	topY := l.HeaderH + l.Gap
	col := func(i int) float64 { return x0 + float64(i)*(l.CardW+l.Gap) }

	l.Stock = Rect{col(0), topY, l.CardW, l.CardH}
	l.Waste = Rect{col(1), topY, l.CardW, l.CardH}
	for i := 0; i < 4; i++ {
		l.Foundations[i] = Rect{col(3 + i), topY, l.CardW, l.CardH}
	}
	tabY := topY + l.CardH + l.Gap*1.5
	for i := 0; i < 7; i++ {
		l.Tableau[i] = Rect{col(i), tabY, l.CardW, l.CardH}
	}
	l.Bottom = h - l.Gap
	return l
}

// PileRect は置き場の先頭カード位置（空のときの枠）を返す。
func (l *Layout) PileRect(id klondike.PileID) Rect {
	switch id.Kind {
	case klondike.Stock:
		return l.Stock
	case klondike.Waste:
		return l.Waste
	case klondike.Foundation:
		return l.Foundations[id.Index]
	case klondike.Tableau:
		return l.Tableau[id.Index]
	}
	return Rect{}
}

// TableauOffsets は列 pile を描画するときの裏向き・表向きのオフセットを返す。
// 列が画面下端を超える場合は圧縮する。
func (l *Layout) TableauOffsets(pile []klondike.Card) (down, up float64) {
	down, up = l.DownOffset, l.UpOffset
	if len(pile) <= 1 {
		return
	}
	var natural float64
	for _, c := range pile[:len(pile)-1] {
		if c.FaceUp {
			natural += up
		} else {
			natural += down
		}
	}
	avail := l.Bottom - l.Tableau[0].Y - l.CardH
	if natural > avail && natural > 0 {
		f := avail / natural
		if f < 0.1 {
			f = 0.1
		}
		down = math.Floor(down * f)
		up = math.Floor(up * f)
		if down < 2 {
			down = 2
		}
		if up < 6 {
			up = 6
		}
	}
	return
}

// CardRect は置き場 id の index 番目（0 が一番下）のカード矩形を返す。
func (l *Layout) CardRect(id klondike.PileID, pile []klondike.Card, index int) Rect {
	base := l.PileRect(id)
	if id.Kind != klondike.Tableau {
		return base
	}
	down, up := l.TableauOffsets(pile)
	y := base.Y
	for i := 0; i < index && i < len(pile); i++ {
		if pile[i].FaceUp {
			y += up
		} else {
			y += down
		}
	}
	return Rect{base.X, y, l.CardW, l.CardH}
}

// TopRect は置き場の一番上のカード（または空枠）の矩形を返す。
func (l *Layout) TopRect(id klondike.PileID, pile []klondike.Card) Rect {
	if len(pile) == 0 {
		return l.PileRect(id)
	}
	return l.CardRect(id, pile, len(pile)-1)
}

// NextRect は置き場に新しくカードを置いたときの矩形を返す。
func (l *Layout) NextRect(id klondike.PileID, pile []klondike.Card) Rect {
	if id.Kind != klondike.Tableau || len(pile) == 0 {
		return l.PileRect(id)
	}
	top := l.CardRect(id, pile, len(pile)-1)
	_, up := l.TableauOffsets(pile)
	top.Y += up
	return top
}

// DropRect はドロップ判定に使う置き場全体の矩形を返す。
func (l *Layout) DropRect(id klondike.PileID, pile []klondike.Card) Rect {
	base := l.PileRect(id)
	if id.Kind != klondike.Tableau || len(pile) == 0 {
		return base
	}
	top := l.CardRect(id, pile, len(pile)-1)
	base.H = top.Y + l.CardH - base.Y
	return base
}
