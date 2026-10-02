// Command cardgen はトランプのカード画像（52 枚 + 裏面）を標準ライブラリと
// golang.org/x/image だけで生成し、assets/cards に配置する。
//
//	go run ./tools/cardgen -out assets/cards -width 180 -height 255
//
// スートマークはベジェ曲線／多角形で描き、ランク文字は Go フォント（gobold）で描く。
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"

	"github.com/nasuton/Solitaire/internal/cardfiles"
	"github.com/nasuton/Solitaire/internal/klondike"
)

var (
	colWhite  = color.RGBA{0xff, 0xff, 0xff, 0xff}
	colBorder = color.RGBA{0x9c, 0xa3, 0x9c, 0xff}
	colRed    = color.RGBA{0xd2, 0x26, 0x2e, 0xff}
	colBlack  = color.RGBA{0x1c, 0x1c, 0x22, 0xff}
	colBackBG = color.RGBA{0x1f, 0x4e, 0x9a, 0xff}
	colBackHL = color.RGBA{0x4f, 0x82, 0xd0, 0xff}
	colBackFR = color.RGBA{0xe8, 0xee, 0xf8, 0xff}
)

func main() {
	out := flag.String("out", filepath.Join("assets", "cards"), "出力先フォルダ")
	width := flag.Int("width", 180, "カード幅（px）")
	height := flag.Int("height", 255, "カード高さ（px）")
	flag.Parse()

	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatal(err)
	}
	ttf, err := opentype.Parse(gobold.TTF)
	if err != nil {
		log.Fatal(err)
	}
	g := &generator{w: *width, h: *height, ttf: ttf}

	var total int64
	count := 0
	for _, s := range klondike.Suits {
		for r := klondike.Rank(1); r <= 13; r++ {
			n, err := writePNG(filepath.Join(*out, cardfiles.Name(s, r)), g.face(s, r))
			if err != nil {
				log.Fatal(err)
			}
			total += n
			count++
		}
	}
	n, err := writePNG(filepath.Join(*out, cardfiles.BackName), g.back())
	if err != nil {
		log.Fatal(err)
	}
	total += n
	count++
	fmt.Printf("%d files, %d bytes (%.1f KiB) -> %s\n", count, total, float64(total)/1024, *out)
}

func writePNG(path string, img image.Image) (int64, error) {
	f, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(f, img); err != nil {
		f.Close()
		return 0, err
	}
	if err := f.Close(); err != nil {
		return 0, err
	}
	st, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return st.Size(), nil
}

type generator struct {
	w, h int
	ttf  *opentype.Font
}

// unit はカード幅に対する比率を px に変換する。
func (g *generator) unit(f float64) float64 { return float64(g.w) * f }

func (g *generator) newFace(size float64) font.Face {
	face, err := opentype.NewFace(g.ttf, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingNone})
	if err != nil {
		log.Fatal(err)
	}
	return face
}

// base は白地・角丸・細枠の下地を描く。
func (g *generator) base() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, g.w, g.h))
	radius := g.unit(0.075)
	border := 1.5
	fillRoundRect(img, 0, 0, float64(g.w), float64(g.h), radius, colBorder)
	fillRoundRect(img, border, border, float64(g.w)-2*border, float64(g.h)-2*border, radius-border, colWhite)
	return img
}

func suitColor(s klondike.Suit) color.RGBA {
	if s.Color() == klondike.Red {
		return colRed
	}
	return colBlack
}

func rankLabel(r klondike.Rank) string {
	switch r {
	case 1:
		return "A"
	case 11:
		return "J"
	case 12:
		return "Q"
	case 13:
		return "K"
	}
	return fmt.Sprint(int(r))
}

// face は表面のカードを描く。
func (g *generator) face(s klondike.Suit, r klondike.Rank) *image.RGBA {
	img := g.base()
	clr := suitColor(s)
	label := rankLabel(r)

	// 左上の索引（ランク文字 + 小さなスート）を別画像に描き、右下には 180° 回転して配置する。
	iw, ih := int(g.unit(0.26)), int(g.unit(0.40))
	idx := image.NewRGBA(image.Rect(0, 0, iw, ih))
	rankSize := g.unit(0.19)
	rankH := rankSize * 0.74 // 大文字・数字の概算高さ
	drawTextCentered(idx, g.newFace(rankSize), label, float64(iw)/2, rankH+g.unit(0.01), clr)
	smallSuit := g.unit(0.15)
	drawSuit(idx, s, (float64(iw)-smallSuit)/2, rankH+g.unit(0.05), smallSuit, clr)

	margin := int(g.unit(0.045))
	draw.Draw(img, image.Rect(margin, margin, margin+iw, margin+ih), idx, image.Point{}, draw.Over)
	rot := rotate180(idx)
	draw.Draw(img, image.Rect(g.w-margin-iw, g.h-margin-ih, g.w-margin, g.h-margin), rot, image.Point{}, draw.Over)

	cx, cy := float64(g.w)/2, float64(g.h)/2
	switch {
	case r >= 11:
		// 絵札: 中央に大きなランク文字、その下にスート。
		bigSize := g.unit(0.56)
		bigH := bigSize * 0.74
		drawTextCentered(img, g.newFace(bigSize), label, cx, cy+bigH*0.05, clr)
		ss := g.unit(0.22)
		drawSuit(img, s, cx-ss/2, cy+bigH*0.45, ss, clr)
	case r == 1:
		size := g.unit(0.56)
		drawSuit(img, s, cx-size/2, cy-size/2, size, clr)
	default:
		size := g.unit(0.46)
		drawSuit(img, s, cx-size/2, cy-size/2, size, clr)
	}
	return img
}

// back は裏面（青系の斜め格子）を描く。
func (g *generator) back() *image.RGBA {
	img := g.base()
	inset := int(g.unit(0.06))
	inner := image.Rect(inset, inset, g.w-inset, g.h-inset)

	pat := image.NewRGBA(image.Rect(0, 0, inner.Dx(), inner.Dy()))
	draw.Draw(pat, pat.Bounds(), image.NewUniform(colBackBG), image.Point{}, draw.Src)
	pw, ph := float64(pat.Bounds().Dx()), float64(pat.Bounds().Dy())
	step := g.unit(0.085)
	band := g.unit(0.014)
	for x := -ph; x < pw+ph; x += step {
		// 右下がりと右上がりの帯を描いて格子にする。
		fillQuad(pat, x, 0, x+band, 0, x+band+ph, ph, x+ph, ph, colBackHL)
		fillQuad(pat, x, ph, x+band, ph, x+band+ph, 0, x+ph, 0, colBackHL)
	}
	draw.Draw(img, inner, pat, image.Point{}, draw.Src)

	// 内側に細い明色の枠。
	fw := g.unit(0.012)
	strokeRect(img, float64(inner.Min.X)+fw, float64(inner.Min.Y)+fw, float64(inner.Dx())-2*fw, float64(inner.Dy())-2*fw, fw, colBackFR)
	return img
}

// ---- 描画ヘルパー ----

// path は vector.Rasterizer に対して、単位座標 → px 変換付きで線分・ベジェを追加する。
type path struct {
	r          *vector.Rasterizer
	ox, oy     float64
	sx, sy     float64
	rotateHalf bool // 単位正方形内で 180° 回転する
}

func (p *path) pt(x, y float64) (float32, float32) {
	if p.rotateHalf {
		x, y = 1-x, 1-y
	}
	return float32(p.ox + x*p.sx), float32(p.oy + y*p.sy)
}
func (p *path) M(x, y float64) { p.r.MoveTo(p.pt(x, y)) }
func (p *path) L(x, y float64) { p.r.LineTo(p.pt(x, y)) }
func (p *path) C(x1, y1, x2, y2, x, y float64) {
	ax, ay := p.pt(x1, y1)
	bx, by := p.pt(x2, y2)
	cx, cy := p.pt(x, y)
	p.r.CubeTo(ax, ay, bx, by, cx, cy)
}
func (p *path) Z() { p.r.ClosePath() }

func newRasterizer(dst *image.RGBA) *vector.Rasterizer {
	r := vector.NewRasterizer(dst.Bounds().Dx(), dst.Bounds().Dy())
	r.DrawOp = draw.Over
	return r
}

func fill(dst *image.RGBA, r *vector.Rasterizer, clr color.Color) {
	r.Draw(dst, dst.Bounds(), image.NewUniform(clr), image.Point{})
}

const kappa = 0.5522847498

func fillRoundRect(dst *image.RGBA, x, y, w, h, rad float64, clr color.Color) {
	r := newRasterizer(dst)
	p := &path{r: r, sx: 1, sy: 1}
	k := rad * kappa
	p.M(x+rad, y)
	p.L(x+w-rad, y)
	p.C(x+w-rad+k, y, x+w, y+rad-k, x+w, y+rad)
	p.L(x+w, y+h-rad)
	p.C(x+w, y+h-rad+k, x+w-rad+k, y+h, x+w-rad, y+h)
	p.L(x+rad, y+h)
	p.C(x+rad-k, y+h, x, y+h-rad+k, x, y+h-rad)
	p.L(x, y+rad)
	p.C(x, y+rad-k, x+rad-k, y, x+rad, y)
	p.Z()
	fill(dst, r, clr)
}

func fillQuad(dst *image.RGBA, x1, y1, x2, y2, x3, y3, x4, y4 float64, clr color.Color) {
	r := newRasterizer(dst)
	p := &path{r: r, sx: 1, sy: 1}
	p.M(x1, y1)
	p.L(x2, y2)
	p.L(x3, y3)
	p.L(x4, y4)
	p.Z()
	fill(dst, r, clr)
}

func strokeRect(dst *image.RGBA, x, y, w, h, t float64, clr color.Color) {
	fillQuad(dst, x, y, x+w, y, x+w, y+t, x, y+t, clr)
	fillQuad(dst, x, y+h-t, x+w, y+h-t, x+w, y+h, x, y+h, clr)
	fillQuad(dst, x, y, x+t, y, x+t, y+h, x, y+h, clr)
	fillQuad(dst, x+w-t, y, x+w, y, x+w, y+h, x+w-t, y+h, clr)
}

// drawSuit はスートマークを (x, y) を左上とする size×size の枠内に描く。
func drawSuit(dst *image.RGBA, s klondike.Suit, x, y, size float64, clr color.Color) {
	switch s {
	case klondike.Hearts:
		r := newRasterizer(dst)
		heart(&path{r: r, ox: x, oy: y + size*0.05, sx: size, sy: size * 0.92})
		fill(dst, r, clr)
	case klondike.Diamonds:
		r := newRasterizer(dst)
		p := &path{r: r, ox: x + size*0.1, oy: y, sx: size * 0.8, sy: size}
		p.M(0.5, 0)
		p.C(0.72, 0.28, 0.88, 0.42, 1, 0.5)
		p.C(0.88, 0.58, 0.72, 0.72, 0.5, 1)
		p.C(0.28, 0.72, 0.12, 0.58, 0, 0.5)
		p.C(0.12, 0.42, 0.28, 0.28, 0.5, 0)
		p.Z()
		fill(dst, r, clr)
	case klondike.Spades:
		r := newRasterizer(dst)
		heart(&path{r: r, ox: x, oy: y, sx: size, sy: size * 0.78, rotateHalf: true})
		fill(dst, r, clr)
		stem(dst, x, y, size, 0.62, clr)
	case klondike.Clubs:
		rad := 0.27
		for _, c := range [][2]float64{{0.5, 0.29}, {0.26, 0.6}, {0.74, 0.6}} {
			r := newRasterizer(dst)
			circle(&path{r: r, ox: x, oy: y, sx: size, sy: size}, c[0], c[1], rad)
			fill(dst, r, clr)
		}
		stem(dst, x, y, size, 0.58, clr)
	}
}

// heart は単位正方形内にハート形を描く（rotateHalf=true で逆さ＝スペード本体）。
func heart(p *path) {
	p.M(0.5, 1)
	p.C(0.5, 1, 0, 0.66, 0, 0.32)
	p.C(0, 0.13, 0.13, 0, 0.28, 0)
	p.C(0.39, 0, 0.47, 0.07, 0.5, 0.17)
	p.C(0.53, 0.07, 0.61, 0, 0.72, 0)
	p.C(0.87, 0, 1, 0.13, 1, 0.32)
	p.C(1, 0.66, 0.5, 1, 0.5, 1)
	p.Z()
}

func circle(p *path, cx, cy, r float64) {
	k := r * kappa
	p.M(cx+r, cy)
	p.C(cx+r, cy+k, cx+k, cy+r, cx, cy+r)
	p.C(cx-k, cy+r, cx-r, cy+k, cx-r, cy)
	p.C(cx-r, cy-k, cx-k, cy-r, cx, cy-r)
	p.C(cx+k, cy-r, cx+r, cy-k, cx+r, cy)
	p.Z()
}

// stem はスペード／クラブの脚を描く（top は脚の付け根の単位 y 座標）。
func stem(dst *image.RGBA, x, y, size, top float64, clr color.Color) {
	r := newRasterizer(dst)
	p := &path{r: r, ox: x, oy: y, sx: size, sy: size}
	p.M(0.5, top)
	p.C(0.5, top+0.2, 0.56, 0.92, 0.7, 1)
	p.L(0.3, 1)
	p.C(0.44, 0.92, 0.5, top+0.2, 0.5, top)
	p.Z()
	fill(dst, r, clr)
}

// drawTextCentered は (cx, baseline) を中心にして文字列を描く。
func drawTextCentered(dst *image.RGBA, face font.Face, s string, cx, baseline float64, clr color.Color) {
	d := &font.Drawer{Dst: dst, Src: image.NewUniform(clr), Face: face}
	adv := d.MeasureString(s)
	d.Dot = fixed.Point26_6{
		X: fixed.Int26_6(cx*64) - adv/2,
		Y: fixed.Int26_6(baseline * 64),
	}
	d.DrawString(s)
}

func rotate180(src *image.RGBA) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(b.Max.X-1-(x-b.Min.X), b.Max.Y-1-(y-b.Min.Y), src.At(x, y))
		}
	}
	return dst
}
