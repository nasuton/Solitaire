package ui

import (
	"fmt"
	"image/color"
	"math"
	"math/rand/v2"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/nasuton/Solitaire/internal/assets"
	"github.com/nasuton/Solitaire/internal/klondike"
	"github.com/nasuton/Solitaire/internal/platform"
)

const (
	tps              = 60
	moveAnimTicks    = 10
	drawAnimTicks    = 8
	autoAnimTicks    = 7
	snapAnimTicks    = 12
	doubleClickTicks = 22 // 約 370ms
	dragThreshold    = 4.0
)

var (
	feltColor   = color.RGBA{0x1d, 0x6b, 0x3a, 0xff}
	headerColor = color.RGBA{0x14, 0x4a, 0x29, 0xff}
	slotColor   = color.RGBA{0xff, 0xff, 0xff, 0x50}
	textColor   = color.RGBA{0xf4, 0xf4, 0xee, 0xff}
	dimColor    = color.RGBA{0xc0, 0xd0, 0xc0, 0xff}
	accentColor = color.RGBA{0xff, 0xe0, 0x70, 0xff}
)

// drag はドラッグ中の状態。
type drag struct {
	from    klondike.PileID
	count   int
	cards   []klondike.Card
	offX    float64 // ポインタとカード左上の差
	offY    float64
	x, y    float64 // 先頭カードの左上
	startX  float64
	startY  float64
	moved   bool
	origin  Rect
	spacing float64
}

// Game は ebiten.Game を実装するソリティア本体。
type Game struct {
	g      *klondike.Game
	cards  *assets.Cards
	layout Layout
	ptr    pointer

	drag   *drag
	anims  []*cardAnim
	hidden map[klondike.PileID]int

	autoRunning bool

	won       bool
	win       *winAnim
	winSettle bool

	started   bool
	startTime time.Time
	finalTime time.Duration

	tick      int
	lastPress struct {
		tick int
		pile klondike.PileID
		x, y float64
	}
	pressedStock bool

	buttons []*button
	rng     *rand.Rand

	lastW, lastH int
	logicalW     int
	logicalH     int
}

// New は seed のゲームを持つ UI を生成する。
func New(cards *assets.Cards, seed int64) *Game {
	ui := &Game{
		cards:  cards,
		hidden: map[klondike.PileID]int{},
		rng:    rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 1)),
	}
	ui.g = klondike.New(seed)
	ui.layout = NewLayout(1280, 800)
	ui.lastPress.tick = -1000
	ui.buttons = []*button{
		{Label: "新規 (N)", OnClick: ui.newGame},
		{Label: "やり直し (R)", OnClick: ui.restart},
		{Label: "戻す (Ctrl+Z)", Enabled: func() bool { return ui.g.CanUndo() }, OnClick: ui.undo},
		{Label: "自動完了 (A)", Enabled: func() bool { return ui.g.CanAutoComplete() && !ui.autoRunning }, OnClick: ui.startAutoComplete},
	}
	platform.RegisterDebugAPI(ui.debugState)
	return ui
}

// Seed は現在のゲームの seed。
func (ui *Game) Seed() int64 { return ui.g.Seed }

// Layout は ebiten.Game の実装。横幅は 1280（縦長画面では 800）に固定し、
// 高さは画面比率に合わせる（極端な比率ではレターボックスになる）。
func (ui *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if outsideWidth <= 0 || outsideHeight <= 0 {
		return 1280, 800
	}
	if outsideWidth != ui.lastW || outsideHeight != ui.lastH {
		ui.lastW, ui.lastH = outsideWidth, outsideHeight
		w := 1280.0
		if outsideHeight > outsideWidth {
			w = 800
		}
		h := math.Round(w * float64(outsideHeight) / float64(outsideWidth))
		h = math.Max(560, math.Min(2400, h))
		ui.logicalW, ui.logicalH = int(w), int(h)
	}
	return ui.logicalW, ui.logicalH
}

// Update は ebiten.Game の実装。
func (ui *Game) Update() error {
	ui.tick++
	if ui.tick == 1 {
		platform.ConfigureCanvas("ソリティアのゲーム画面。マウスまたはタッチでカードを操作します。")
	}
	if ui.logicalW > 0 && (float64(ui.logicalW) != ui.layout.W || float64(ui.logicalH) != ui.layout.H) {
		ui.layout = NewLayout(float64(ui.logicalW), float64(ui.logicalH))
	}
	ui.layoutButtons()
	ui.ptr.update()

	ui.handleKeys()

	clicked := false
	for _, b := range ui.buttons {
		if b.update(&ui.ptr) {
			clicked = true
		}
	}
	if !clicked {
		ui.handlePointer()
	}

	ui.updateAnims()
	ui.updateAutoComplete()
	ui.checkWin()
	if ui.win != nil {
		if ui.win.update(&ui.layout) {
			ui.winSettle = true
		}
	}
	ui.rebuildHidden()
	return nil
}

func (ui *Game) handleKeys() {
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyN):
		ui.newGame()
	case inpututil.IsKeyJustPressed(ebiten.KeyR):
		ui.restart()
	case ctrl && inpututil.IsKeyJustPressed(ebiten.KeyZ):
		ui.undo()
	case inpututil.IsKeyJustPressed(ebiten.KeySpace):
		if ui.drag == nil && !ui.autoRunning && !ui.won {
			ui.drawStock()
		}
	case inpututil.IsKeyJustPressed(ebiten.KeyA):
		if ui.g.CanAutoComplete() {
			ui.startAutoComplete()
		}
	}
}

// ---- ゲーム操作 ----

func (ui *Game) resetState() {
	ui.drag = nil
	ui.anims = nil
	ui.autoRunning = false
	ui.won = false
	ui.win = nil
	ui.winSettle = false
	ui.started = false
	ui.finalTime = 0
	ui.pressedStock = false
}

func (ui *Game) newGame() {
	seed := ui.rng.Int64N(1_000_000)
	ui.g = klondike.New(seed)
	ui.resetState()
	platform.SetSeedInURL(seed)
}

func (ui *Game) restart() {
	ui.g.Restart()
	ui.resetState()
}

func (ui *Game) undo() {
	if !ui.g.CanUndo() {
		return
	}
	ui.finishAnims()
	ui.drag = nil
	ui.autoRunning = false
	if err := ui.g.Undo(); err != nil {
		return
	}
	if ui.won {
		// 勝利状態から戻した場合はタイマーを再開する。
		ui.won = false
		ui.win = nil
		ui.winSettle = false
		ui.startTime = time.Now().Add(-ui.finalTime)
	}
}

func (ui *Game) startAutoComplete() {
	if !ui.g.CanAutoComplete() {
		return
	}
	ui.drag = nil
	ui.autoRunning = true
	ui.markStarted()
}

func (ui *Game) markStarted() {
	if !ui.started {
		ui.started = true
		ui.startTime = time.Now()
	}
}

// drawStock は山札をめくる（または再循環する）。
func (ui *Game) drawStock() {
	if !ui.g.CanDraw() {
		return
	}
	ui.finishAnims()
	recycle := ui.g.WillRecycle()
	if err := ui.g.Draw(); err != nil {
		return
	}
	ui.markStarted()
	if recycle {
		return
	}
	top, _ := ui.g.Top(klondike.WasteID)
	ui.anims = append(ui.anims, &cardAnim{
		cards:     []klondike.Card{top},
		fromX:     ui.layout.Stock.X,
		fromY:     ui.layout.Stock.Y,
		toX:       ui.layout.Waste.X,
		toY:       ui.layout.Waste.Y,
		dur:       drawAnimTicks,
		hide:      klondike.WasteID,
		hideCount: 1,
	})
}

// applyMove は手を適用し、fromX/fromY から移動先へ飛ぶアニメーションを追加する。
func (ui *Game) applyMove(m klondike.Move, fromX, fromY float64, dur int) bool {
	dest := ui.layout.NextRect(m.To, ui.g.Pile(m.To))
	if err := ui.g.Apply(m); err != nil {
		return false
	}
	ui.markStarted()
	to := ui.g.Pile(m.To)
	moved := append([]klondike.Card(nil), to[len(to)-m.Count:]...)
	_, up := ui.layout.TableauOffsets(to)
	spacing := 0.0
	if m.To.Kind == klondike.Tableau {
		spacing = up
	}
	ui.anims = append(ui.anims, &cardAnim{
		cards:     moved,
		fromX:     fromX,
		fromY:     fromY,
		toX:       dest.X,
		toY:       dest.Y,
		spacing:   spacing,
		dur:       dur,
		hide:      m.To,
		hideCount: m.Count,
	})
	return true
}

// ---- ポインタ処理 ----

func (ui *Game) handlePointer() {
	p := &ui.ptr
	l := &ui.layout

	if ui.drag != nil {
		d := ui.drag
		d.x = p.X - d.offX
		d.y = p.Y - d.offY
		if math.Abs(p.X-d.startX) > dragThreshold || math.Abs(p.Y-d.startY) > dragThreshold {
			d.moved = true
		}
		if p.JustReleased || !p.Pressed {
			ui.endDrag()
		}
		return
	}

	if ui.autoRunning || ui.won {
		return
	}

	if p.JustPressed {
		if l.Stock.Contains(p.X, p.Y) {
			ui.pressedStock = true
		} else {
			ui.beginPress(p.X, p.Y)
			return
		}
	}
	// 押下と離しが同じ tick に起きる高速タップも拾う。
	if p.JustReleased && ui.pressedStock {
		ui.pressedStock = false
		if l.Stock.Contains(p.X, p.Y) {
			ui.drawStock()
		}
	}
}

// beginPress はカード上での押下を処理する（ダブルクリック判定とドラッグ開始）。
func (ui *Game) beginPress(x, y float64) {
	id, index, ok := ui.hitCard(x, y)
	if !ok {
		ui.lastPress.tick = -1000
		return
	}
	pile := ui.g.Pile(id)
	isDouble := ui.tick-ui.lastPress.tick <= doubleClickTicks &&
		ui.lastPress.pile == id &&
		math.Abs(ui.lastPress.x-x) < 12 && math.Abs(ui.lastPress.y-y) < 12
	ui.lastPress.tick = ui.tick
	ui.lastPress.pile = id
	ui.lastPress.x, ui.lastPress.y = x, y

	if isDouble && index == len(pile)-1 {
		if m, ok := ui.g.FoundationMoveFor(id); ok {
			ui.finishAnims()
			src := ui.layout.CardRect(id, pile, index)
			ui.applyMove(m, src.X, src.Y, moveAnimTicks)
			ui.lastPress.tick = -1000
			return
		}
	}

	count := len(pile) - index
	if id.Kind != klondike.Tableau && count != 1 {
		return
	}
	ui.finishAnims()
	src := ui.layout.CardRect(id, pile, index)
	_, up := ui.layout.TableauOffsets(pile)
	spacing := ui.layout.UpOffset
	if id.Kind == klondike.Tableau {
		spacing = up
	}
	ui.drag = &drag{
		from:    id,
		count:   count,
		cards:   append([]klondike.Card(nil), pile[index:]...),
		offX:    x - src.X,
		offY:    y - src.Y,
		x:       src.X,
		y:       src.Y,
		startX:  x,
		startY:  y,
		origin:  src,
		spacing: spacing,
	}
}

// hitCard は点 (x, y) にあるドラッグ可能なカードを返す。
func (ui *Game) hitCard(x, y float64) (klondike.PileID, int, bool) {
	l := &ui.layout
	if l.Waste.Contains(x, y) && len(ui.g.Waste) > 0 {
		return klondike.WasteID, len(ui.g.Waste) - 1, true
	}
	for i := 0; i < 4; i++ {
		if l.Foundations[i].Contains(x, y) && len(ui.g.Foundations[i]) > 0 {
			return klondike.FoundationID(i), len(ui.g.Foundations[i]) - 1, true
		}
	}
	for i := 0; i < 7; i++ {
		id := klondike.TableauID(i)
		pile := ui.g.Tableau[i]
		for idx := len(pile) - 1; idx >= 0; idx-- {
			if l.CardRect(id, pile, idx).Contains(x, y) {
				if !pile[idx].FaceUp {
					return id, idx, false
				}
				return id, idx, true
			}
		}
	}
	return klondike.PileID{}, 0, false
}

// endDrag はドロップ先を決めて移動またはスナップバックする。
func (ui *Game) endDrag() {
	d := ui.drag
	ui.drag = nil
	if !d.moved {
		return
	}
	cardRect := Rect{d.x, d.y, ui.layout.CardW, ui.layout.CardH}

	var best klondike.PileID
	bestArea := 0.0
	candidates := make([]klondike.PileID, 0, 11)
	for i := 0; i < 4; i++ {
		candidates = append(candidates, klondike.FoundationID(i))
	}
	for i := 0; i < 7; i++ {
		candidates = append(candidates, klondike.TableauID(i))
	}
	for _, id := range candidates {
		if id == d.from {
			continue
		}
		area := cardRect.OverlapArea(ui.layout.DropRect(id, ui.g.Pile(id)))
		if area > bestArea {
			bestArea = area
			best = id
		}
	}
	if bestArea > 0 {
		m := klondike.Move{From: d.from, To: best, Count: d.count}
		if ui.g.CanMove(m) && ui.applyMove(m, d.x, d.y, moveAnimTicks) {
			return
		}
	}
	// スナップバック
	ui.anims = append(ui.anims, &cardAnim{
		cards:     d.cards,
		fromX:     d.x,
		fromY:     d.y,
		toX:       d.origin.X,
		toY:       d.origin.Y,
		spacing:   d.spacing,
		dur:       snapAnimTicks,
		hide:      d.from,
		hideCount: d.count,
	})
}

// ---- アニメーション ----

func (ui *Game) updateAnims() {
	n := 0
	for _, a := range ui.anims {
		a.t++
		if a.done() {
			if a.onDone != nil {
				a.onDone()
			}
			continue
		}
		ui.anims[n] = a
		n++
	}
	ui.anims = ui.anims[:n]
}

// finishAnims は進行中のアニメーションを即座に完了させる。
func (ui *Game) finishAnims() {
	for _, a := range ui.anims {
		if a.onDone != nil {
			a.onDone()
		}
	}
	ui.anims = ui.anims[:0]
}

func (ui *Game) updateAutoComplete() {
	if !ui.autoRunning || len(ui.anims) > 0 {
		return
	}
	m, ok := ui.g.NextAutoCompleteMove()
	if !ok {
		ui.autoRunning = false
		return
	}
	src := ui.layout.TopRect(m.From, ui.g.Pile(m.From))
	ui.applyMove(m, src.X, src.Y, autoAnimTicks)
}

func (ui *Game) checkWin() {
	if ui.won || len(ui.anims) > 0 || !ui.g.IsWon() {
		return
	}
	ui.won = true
	ui.autoRunning = false
	ui.finalTime = ui.elapsed()
	ui.win = newWinAnim(ui.g, ui.g.Seed)
}

func (ui *Game) rebuildHidden() {
	for k := range ui.hidden {
		delete(ui.hidden, k)
	}
	if ui.drag != nil {
		ui.hidden[ui.drag.from] += ui.drag.count
	}
	for _, a := range ui.anims {
		ui.hidden[a.hide] += a.hideCount
	}
	if ui.win != nil {
		for i, n := range ui.win.taken {
			ui.hidden[klondike.FoundationID(i)] += n
		}
	}
}

func (ui *Game) elapsed() time.Duration {
	if ui.won {
		return ui.finalTime
	}
	if !ui.started {
		return 0
	}
	return time.Since(ui.startTime)
}

// ---- 描画 ----

// Draw は ebiten.Game の実装。
func (ui *Game) Draw(screen *ebiten.Image) {
	l := &ui.layout
	screen.Fill(feltColor)
	vector.DrawFilledRect(screen, 0, 0, float32(l.W), float32(l.HeaderH), headerColor, false)

	ui.drawHeader(screen)
	ui.drawPiles(screen)
	for _, a := range ui.anims {
		x, y := a.pos()
		for i, c := range a.cards {
			ui.drawCard(screen, c, x, y+float64(i)*a.spacing, true)
		}
	}
	if ui.drag != nil {
		d := ui.drag
		for i, c := range d.cards {
			ui.drawCard(screen, c, d.x, d.y+float64(i)*d.spacing, true)
		}
	}
	if ui.win != nil {
		for i := range ui.win.cards {
			fc := &ui.win.cards[i]
			if fc.alive {
				ui.drawCard(screen, fc.card, fc.x, fc.y, false)
			}
		}
	}
	if ui.won {
		ui.drawWinOverlay(screen)
	}
}

func (ui *Game) layoutButtons() {
	l := &ui.layout
	s := l.TextScale
	bh := 13*s + 8
	pad := 10.0
	y := (l.HeaderH - bh) / 2
	if l.TwoRowHeader {
		y = l.HeaderH - bh - 10
	}
	total := 0.0
	for _, b := range ui.buttons {
		w, _ := textSize(b.Label, s)
		b.Rect = Rect{W: w + 2*pad, H: bh}
		total += b.Rect.W + 8
	}
	total -= 8
	x := l.W - l.Gap - total
	if l.TwoRowHeader {
		x = (l.W - total) / 2
	}
	for _, b := range ui.buttons {
		b.Rect.X = x
		b.Rect.Y = y
		x += b.Rect.W + 8
	}
}

func (ui *Game) drawHeader(screen *ebiten.Image) {
	l := &ui.layout
	s := l.TextScale
	e := ui.elapsed()
	stats := fmt.Sprintf("Seed:%d  手数:%d  時間:%02d:%02d  スコア:%d",
		ui.g.Seed, ui.g.Moves, int(e.Minutes()), int(e.Seconds())%60, ui.g.Score)
	_, th := textSize(stats, s)
	y := 10.0
	if !l.TwoRowHeader {
		y = (l.HeaderH - th) / 2
	}
	drawText(screen, stats, l.Gap, y, s, textColor)
	for _, b := range ui.buttons {
		b.draw(screen, s)
	}
}

func (ui *Game) drawPiles(screen *ebiten.Image) {
	l := &ui.layout
	g := ui.g

	// 山札
	ui.drawSlot(screen, l.Stock)
	if n := len(g.Stock) - ui.hidden[klondike.StockID]; n > 0 {
		ui.drawCard(screen, g.Stock[len(g.Stock)-1], l.Stock.X, l.Stock.Y, false)
	} else if len(g.Waste) > 0 {
		drawTextCentered(screen, "↻", l.Stock, l.TextScale*2, slotColor)
	}
	// 捨て札
	ui.drawSlot(screen, l.Waste)
	if n := len(g.Waste) - ui.hidden[klondike.WasteID]; n > 0 {
		ui.drawCard(screen, g.Waste[n-1], l.Waste.X, l.Waste.Y, false)
	}
	// 組札
	for i := 0; i < 4; i++ {
		r := l.Foundations[i]
		ui.drawSlot(screen, r)
		f := g.Foundations[i]
		n := len(f) - ui.hidden[klondike.FoundationID(i)]
		if n > 0 {
			ui.drawCard(screen, f[n-1], r.X, r.Y, false)
		} else {
			drawTextCentered(screen, "A", r, l.TextScale*2, slotColor)
		}
	}
	// 場札
	for i := 0; i < 7; i++ {
		id := klondike.TableauID(i)
		pile := g.Tableau[i]
		ui.drawSlot(screen, l.Tableau[i])
		n := len(pile) - ui.hidden[id]
		for idx := 0; idx < n; idx++ {
			r := l.CardRect(id, pile, idx)
			ui.drawCard(screen, pile[idx], r.X, r.Y, false)
		}
	}
}

func (ui *Game) drawSlot(screen *ebiten.Image, r Rect) {
	vector.StrokeRect(screen, float32(r.X)+1, float32(r.Y)+1, float32(r.W)-2, float32(r.H)-2, 2, slotColor, true)
}

func (ui *Game) drawCard(screen *ebiten.Image, c klondike.Card, x, y float64, shadow bool) {
	img := ui.cards.Image(c)
	sx := ui.layout.CardW / float64(ui.cards.Width)
	sy := ui.layout.CardH / float64(ui.cards.Height)
	if shadow {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(sx, sy)
		op.GeoM.Translate(x+4, y+5)
		op.ColorScale.Scale(0, 0, 0, 0.35)
		op.Filter = ebiten.FilterLinear
		screen.DrawImage(img, op)
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(sx, sy)
	op.GeoM.Translate(x, y)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(img, op)
}

func (ui *Game) drawWinOverlay(screen *ebiten.Image) {
	l := &ui.layout
	s := l.TextScale
	title := "クリア！"
	e := ui.finalTime
	lines := []string{
		fmt.Sprintf("手数 %d   時間 %02d:%02d   スコア %d", ui.g.Moves, int(e.Minutes()), int(e.Seconds())%60, ui.g.Score),
		"N キーまたは「新規」で次のゲーム",
	}
	tw, th := textSize(title, s*3)
	maxW := tw
	for _, ln := range lines {
		w, _ := textSize(ln, s)
		maxW = math.Max(maxW, w)
	}
	_, lh := textSize("A", s)
	pw := maxW + 60
	ph := th + float64(len(lines))*(lh+8) + 60
	px := (l.W - pw) / 2
	py := l.HeaderH + (l.H-l.HeaderH-ph)/2
	vector.DrawFilledRect(screen, float32(px), float32(py), float32(pw), float32(ph), color.RGBA{0, 0, 0, 0xb0}, false)
	vector.StrokeRect(screen, float32(px), float32(py), float32(pw), float32(ph), 2, accentColor, false)
	drawText(screen, title, px+(pw-tw)/2, py+24, s*3, accentColor)
	y := py + 24 + th + 16
	for _, ln := range lines {
		w, _ := textSize(ln, s)
		drawText(screen, ln, px+(pw-w)/2, y, s, textColor)
		y += lh + 8
	}
}

// debugState はブラウザからの検証用に状態を返す。
func (ui *Game) debugState() map[string]any {
	l := &ui.layout
	rect := func(r Rect) map[string]any {
		return map[string]any{"x": r.X, "y": r.Y, "w": r.W, "h": r.H}
	}
	tops := make([]any, 7)
	for i := 0; i < 7; i++ {
		id := klondike.TableauID(i)
		tops[i] = rect(l.TopRect(id, ui.g.Pile(id)))
	}
	foundations := make([]any, 4)
	foundationCount := 0
	for i := 0; i < 4; i++ {
		id := klondike.FoundationID(i)
		foundations[i] = rect(l.PileRect(id))
		foundationCount += len(ui.g.Pile(id))
	}
	return map[string]any{
		"seed":        ui.g.Seed,
		"moves":       ui.g.Moves,
		"score":       ui.g.Score,
		"stock":       len(ui.g.Stock),
		"waste":       len(ui.g.Waste),
		"foundation":  foundationCount,
		"won":         ui.won,
		"canUndo":     ui.g.CanUndo(),
		"logicalW":    l.W,
		"logicalH":    l.H,
		"stockRect":   rect(l.Stock),
		"wasteRect":   rect(l.Waste),
		"undoButton":  rect(ui.buttons[2].Rect),
		"tableauTops": tops,
		"foundations": foundations,
	}
}
