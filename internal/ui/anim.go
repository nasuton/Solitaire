package ui

import (
	"math"
	"math/rand/v2"

	"github.com/nasuton/Solitaire/internal/klondike"
)

// cardAnim は複数枚のカードが from から to へ飛ぶアニメーション。
// 飛行中は hide で示した置き場の上から hideCount 枚を描画しない。
type cardAnim struct {
	cards     []klondike.Card
	fromX     float64
	fromY     float64
	toX       float64
	toY       float64
	spacing   float64 // 連なりを描くときの縦間隔
	t, dur    int     // tick
	hide      klondike.PileID
	hideCount int
	onDone    func()
}

func (a *cardAnim) progress() float64 {
	if a.dur <= 0 {
		return 1
	}
	p := float64(a.t) / float64(a.dur)
	if p > 1 {
		p = 1
	}
	// ease-out cubic
	return 1 - math.Pow(1-p, 3)
}

func (a *cardAnim) pos() (x, y float64) {
	p := a.progress()
	return a.fromX + (a.toX-a.fromX)*p, a.fromY + (a.toY-a.fromY)*p
}

func (a *cardAnim) done() bool { return a.t >= a.dur }

// flyingCard はクリア演出で跳ねるカード。
type flyingCard struct {
	card   klondike.Card
	x, y   float64
	vx, vy float64
	alive  bool
}

// winAnim はクリア時の演出。組札からカードが順に飛び出して床で跳ねる。
type winAnim struct {
	cards    []flyingCard
	launched int
	tick     int
	rng      *rand.Rand
	// 各組札からすでに飛ばした枚数（描画時に隠す）。
	taken [4]int
	// 発射順（組札 index と残り枚数から決める）。
	order []klondike.PileID
}

func newWinAnim(g *klondike.Game, seed int64) *winAnim {
	w := &winAnim{rng: rand.New(rand.NewPCG(uint64(seed), 7))}
	// 4 つの組札から交互に 1 枚ずつ（上から）飛ばす。
	for n := 0; n < 13; n++ {
		for f := 0; f < 4; f++ {
			w.order = append(w.order, klondike.FoundationID(f))
		}
	}
	for _, id := range w.order {
		f := g.Foundations[id.Index]
		w.cards = append(w.cards, flyingCard{card: f[len(f)-1-w.taken[id.Index]]})
		w.taken[id.Index]++
	}
	w.taken = [4]int{}
	return w
}

// update は演出を 1 tick 進める。全カードが画面外に出たら true。
func (w *winAnim) update(l *Layout) bool {
	w.tick++
	if w.launched < len(w.cards) && w.tick%4 == 0 {
		id := w.order[w.launched]
		r := l.Foundations[id.Index]
		fc := &w.cards[w.launched]
		fc.x, fc.y = r.X, r.Y
		fc.vx = (w.rng.Float64()*6 + 3) * l.CardW / 100
		if w.rng.IntN(2) == 0 {
			fc.vx = -fc.vx
		}
		fc.vy = -(w.rng.Float64()*8 + 2) * l.CardH / 100
		fc.alive = true
		w.taken[id.Index]++
		w.launched++
	}
	gravity := 0.45 * l.CardH / 100
	allGone := w.launched == len(w.cards)
	for i := range w.cards {
		fc := &w.cards[i]
		if !fc.alive {
			continue
		}
		fc.vy += gravity
		fc.x += fc.vx
		fc.y += fc.vy
		if fc.y+l.CardH > l.H {
			fc.y = l.H - l.CardH
			fc.vy = -fc.vy * 0.78
		}
		if fc.x+l.CardW < 0 || fc.x > l.W {
			fc.alive = false
			continue
		}
		allGone = false
	}
	return allGone
}
