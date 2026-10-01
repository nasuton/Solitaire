package klondike

import (
	"errors"
	"fmt"
	"math/rand/v2"
)

// PileKind は札の置き場の種類。
type PileKind int

// 置き場の種類。
const (
	Stock      PileKind = iota // 山札
	Waste                      // 捨て札
	Foundation                 // 組札（4 つ）
	Tableau                    // 場札（7 列）
)

// String は PileKind の名前を返す。
func (k PileKind) String() string {
	switch k {
	case Stock:
		return "Stock"
	case Waste:
		return "Waste"
	case Foundation:
		return "Foundation"
	case Tableau:
		return "Tableau"
	}
	return fmt.Sprintf("PileKind(%d)", int(k))
}

// PileID は置き場を一意に識別する。Index は Foundation(0..3) / Tableau(0..6) でのみ意味を持つ。
type PileID struct {
	Kind  PileKind
	Index int
}

// 置き場 ID を生成するヘルパー。
var (
	StockID = PileID{Kind: Stock}
	WasteID = PileID{Kind: Waste}
)

// FoundationID は i 番目の組札の ID を返す。
func FoundationID(i int) PileID { return PileID{Kind: Foundation, Index: i} }

// TableauID は i 番目の場札列の ID を返す。
func TableauID(i int) PileID { return PileID{Kind: Tableau, Index: i} }

// Valid は ID が有効な置き場を指すかを返す。
func (p PileID) Valid() bool {
	switch p.Kind {
	case Stock, Waste:
		return true
	case Foundation:
		return p.Index >= 0 && p.Index < 4
	case Tableau:
		return p.Index >= 0 && p.Index < 7
	}
	return false
}

// String は PileID の表記を返す。
func (p PileID) String() string {
	switch p.Kind {
	case Foundation, Tableau:
		return fmt.Sprintf("%s[%d]", p.Kind, p.Index)
	}
	return p.Kind.String()
}

// Move はカード移動を表す。From の末尾 Count 枚を To の末尾へ移す。
type Move struct {
	From  PileID
	To    PileID
	Count int
}

// スコア定数（Windows 標準ソリティア方式）。
const (
	ScoreWasteToTableau      = 5
	ScoreTableauToFoundation = 10
	ScoreWasteToFoundation   = 10
	ScoreFlipTableau         = 5
	ScoreFoundationToTableau = -15
	ScoreRecycle             = -100
)

// エラー定義。
var (
	ErrInvalidMove   = errors.New("klondike: invalid move")
	ErrNothingToDo   = errors.New("klondike: nothing to do")
	ErrInvalidPile   = errors.New("klondike: invalid pile")
	ErrNothingToUndo = errors.New("klondike: nothing to undo")
)

// Game はクロンダイク 1 ゲームの状態。フィールドは描画のために公開しているが、
// 変更は Game のメソッド経由で行うこと。
type Game struct {
	Seed        int64
	Stock       []Card
	Waste       []Card
	Foundations [4][]Card
	Tableau     [7][]Card

	Moves    int // 手数
	Score    int // スコア（下限 0）
	Recycles int // 山札の再循環回数

	history []snapshot
}

// snapshot は Undo 用の状態複製。
type snapshot struct {
	stock       []Card
	waste       []Card
	foundations [4][]Card
	tableau     [7][]Card
	moves       int
	score       int
	recycles    int
}

// New は seed に基づいてシャッフル・配牌した新しいゲームを返す。
// 同じ seed なら常に同じ配牌になる。
func New(seed int64) *Game {
	g := &Game{Seed: seed}
	g.deal()
	return g
}

// Restart は同じ seed で最初から配り直す。
func (g *Game) Restart() {
	*g = Game{Seed: g.Seed}
	g.deal()
}

func (g *Game) deal() {
	deck := NewDeck()
	r := rand.New(rand.NewPCG(uint64(g.Seed), 0x5a5a5a5a))
	r.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

	idx := 0
	for col := 0; col < 7; col++ {
		g.Tableau[col] = make([]Card, 0, 20)
		for n := 0; n <= col; n++ {
			c := deck[idx]
			idx++
			c.FaceUp = n == col
			g.Tableau[col] = append(g.Tableau[col], c)
		}
	}
	g.Stock = append(make([]Card, 0, 24), deck[idx:]...)
	g.Waste = make([]Card, 0, 24)
	for i := range g.Foundations {
		g.Foundations[i] = make([]Card, 0, 13)
	}
}

// Pile は指定した置き場のカード列を返す（末尾が一番上）。無効な ID なら nil。
func (g *Game) Pile(id PileID) []Card {
	switch id.Kind {
	case Stock:
		return g.Stock
	case Waste:
		return g.Waste
	case Foundation:
		if id.Index >= 0 && id.Index < 4 {
			return g.Foundations[id.Index]
		}
	case Tableau:
		if id.Index >= 0 && id.Index < 7 {
			return g.Tableau[id.Index]
		}
	}
	return nil
}

func (g *Game) pilePtr(id PileID) *[]Card {
	switch id.Kind {
	case Stock:
		return &g.Stock
	case Waste:
		return &g.Waste
	case Foundation:
		if id.Index >= 0 && id.Index < 4 {
			return &g.Foundations[id.Index]
		}
	case Tableau:
		if id.Index >= 0 && id.Index < 7 {
			return &g.Tableau[id.Index]
		}
	}
	return nil
}

// Top は置き場の一番上のカードを返す。空なら ok=false。
func (g *Game) Top(id PileID) (c Card, ok bool) {
	p := g.Pile(id)
	if len(p) == 0 {
		return Card{}, false
	}
	return p[len(p)-1], true
}

// CanDraw は山札めくり（または再循環）が可能かを返す。
func (g *Game) CanDraw() bool {
	return len(g.Stock) > 0 || len(g.Waste) > 0
}

// Draw は山札から 1 枚めくって捨て札に置く。山札が空なら捨て札を裏返して山札に戻す（再循環）。
// 再循環はスコア −100。山札・捨て札とも空なら ErrNothingToDo。
func (g *Game) Draw() error {
	if !g.CanDraw() {
		return ErrNothingToDo
	}
	g.pushHistory()
	if len(g.Stock) == 0 {
		// 捨て札を逆順にして裏向きで山札へ。
		for i := len(g.Waste) - 1; i >= 0; i-- {
			c := g.Waste[i]
			c.FaceUp = false
			g.Stock = append(g.Stock, c)
		}
		g.Waste = g.Waste[:0]
		g.Recycles++
		g.addScore(ScoreRecycle)
	} else {
		c := g.Stock[len(g.Stock)-1]
		g.Stock = g.Stock[:len(g.Stock)-1]
		c.FaceUp = true
		g.Waste = append(g.Waste, c)
	}
	g.Moves++
	return nil
}

// WillRecycle は次の Draw が再循環になるかを返す。
func (g *Game) WillRecycle() bool {
	return len(g.Stock) == 0 && len(g.Waste) > 0
}

// CanMove は m が合法手かを返す。
func (g *Game) CanMove(m Move) bool {
	return g.validate(m) == nil
}

func (g *Game) validate(m Move) error {
	if m.From == m.To {
		return ErrInvalidMove
	}
	if !m.From.Valid() || !m.To.Valid() {
		return ErrInvalidPile
	}
	from := g.Pile(m.From)
	to := g.pilePtr(m.To)
	if m.Count <= 0 || m.Count > len(from) {
		return ErrInvalidMove
	}

	switch m.From.Kind {
	case Stock:
		return ErrInvalidMove // 山札からは Draw のみ
	case Waste, Foundation:
		if m.Count != 1 {
			return ErrInvalidMove
		}
	case Tableau:
		// 移動する連なりは全て表向きで、交互色・降順でなければならない。
		run := from[len(from)-m.Count:]
		if !isValidRun(run) {
			return ErrInvalidMove
		}
	}

	moving := from[len(from)-m.Count:]
	bottom := moving[0]
	if !bottom.FaceUp {
		return ErrInvalidMove
	}

	switch m.To.Kind {
	case Stock, Waste:
		return ErrInvalidMove
	case Foundation:
		if m.Count != 1 {
			return ErrInvalidMove
		}
		if !canPlaceOnFoundation(*to, bottom) {
			return ErrInvalidMove
		}
	case Tableau:
		if !canPlaceOnTableau(*to, bottom) {
			return ErrInvalidMove
		}
	}
	return nil
}

// isValidRun は run が表向きで交互色・1 ずつ降順かを返す。
func isValidRun(run []Card) bool {
	for i, c := range run {
		if !c.FaceUp {
			return false
		}
		if i > 0 {
			prev := run[i-1]
			if prev.Rank != c.Rank+1 || prev.Suit.Color() == c.Suit.Color() {
				return false
			}
		}
	}
	return true
}

func canPlaceOnFoundation(f []Card, c Card) bool {
	if len(f) == 0 {
		return c.Rank == Ace
	}
	top := f[len(f)-1]
	return top.Suit == c.Suit && top.Rank+1 == c.Rank
}

func canPlaceOnTableau(t []Card, c Card) bool {
	if len(t) == 0 {
		return c.Rank == King
	}
	top := t[len(t)-1]
	if !top.FaceUp {
		return false
	}
	return top.Rank == c.Rank+1 && top.Suit.Color() != c.Suit.Color()
}

// Apply は m を適用する。不正な手なら状態を変えずにエラーを返す。
// 移動後に場札の一番上が裏向きなら自動でめくる（+5）。
func (g *Game) Apply(m Move) error {
	if err := g.validate(m); err != nil {
		return err
	}
	g.pushHistory()
	from := g.pilePtr(m.From)
	to := g.pilePtr(m.To)

	moving := (*from)[len(*from)-m.Count:]
	*to = append(*to, moving...)
	*from = (*from)[:len(*from)-m.Count]

	g.addScore(scoreFor(m))

	if m.From.Kind == Tableau && len(*from) > 0 {
		top := &(*from)[len(*from)-1]
		if !top.FaceUp {
			top.FaceUp = true
			g.addScore(ScoreFlipTableau)
		}
	}
	g.Moves++
	return nil
}

func scoreFor(m Move) int {
	switch {
	case m.From.Kind == Waste && m.To.Kind == Tableau:
		return ScoreWasteToTableau
	case m.From.Kind == Waste && m.To.Kind == Foundation:
		return ScoreWasteToFoundation
	case m.From.Kind == Tableau && m.To.Kind == Foundation:
		return ScoreTableauToFoundation
	case m.From.Kind == Foundation && m.To.Kind == Tableau:
		return ScoreFoundationToTableau
	}
	return 0
}

func (g *Game) addScore(delta int) {
	g.Score += delta
	if g.Score < 0 {
		g.Score = 0
	}
}

// FoundationMoveFor は from の一番上のカードを置ける組札への Move を返す。
// 置ける組札が無ければ ok=false。ダブルクリックの自動移動に使う。
func (g *Game) FoundationMoveFor(from PileID) (Move, bool) {
	if from.Kind != Waste && from.Kind != Tableau {
		return Move{}, false
	}
	c, ok := g.Top(from)
	if !ok || !c.FaceUp {
		return Move{}, false
	}
	// 同スートの組札を優先し、A は最初の空き組札へ。
	for i, f := range g.Foundations {
		if len(f) > 0 && canPlaceOnFoundation(f, c) {
			return Move{From: from, To: FoundationID(i), Count: 1}, true
		}
	}
	if c.Rank == Ace {
		for i, f := range g.Foundations {
			if len(f) == 0 {
				return Move{From: from, To: FoundationID(i), Count: 1}, true
			}
		}
	}
	return Move{}, false
}

// IsWon は 4 つの組札がすべて K まで揃ったかを返す。
func (g *Game) IsWon() bool {
	for _, f := range g.Foundations {
		if len(f) != 13 {
			return false
		}
	}
	return true
}

// CanAutoComplete は自動完了が可能か（山札・捨て札が空で場札が全て表向き）を返す。
// 既に勝利している場合は false。
func (g *Game) CanAutoComplete() bool {
	if g.IsWon() {
		return false
	}
	if len(g.Stock) != 0 || len(g.Waste) != 0 {
		return false
	}
	for _, t := range g.Tableau {
		for _, c := range t {
			if !c.FaceUp {
				return false
			}
		}
	}
	return true
}

// NextAutoCompleteMove は自動完了で次に組札へ送る 1 手を返す。
// 場札の一番上のうち、最もランクの低いカードを優先する。無ければ ok=false。
func (g *Game) NextAutoCompleteMove() (Move, bool) {
	if !g.CanAutoComplete() {
		return Move{}, false
	}
	best := Move{}
	bestRank := King + 1
	found := false
	for i := range g.Tableau {
		id := TableauID(i)
		m, ok := g.FoundationMoveFor(id)
		if !ok {
			continue
		}
		c, _ := g.Top(id)
		if c.Rank < bestRank {
			bestRank = c.Rank
			best = m
			found = true
		}
	}
	return best, found
}

// CanUndo は Undo 可能かを返す。
func (g *Game) CanUndo() bool {
	return len(g.history) > 0
}

// Undo は直前の操作（Draw / Apply）を取り消す。手数は巻き戻した時点の値に戻る。
func (g *Game) Undo() error {
	if len(g.history) == 0 {
		return ErrNothingToUndo
	}
	s := g.history[len(g.history)-1]
	g.history = g.history[:len(g.history)-1]
	g.Stock = s.stock
	g.Waste = s.waste
	g.Foundations = s.foundations
	g.Tableau = s.tableau
	g.Moves = s.moves
	g.Score = s.score
	g.Recycles = s.recycles
	return nil
}

func (g *Game) pushHistory() {
	s := snapshot{
		stock:    cloneCards(g.Stock),
		waste:    cloneCards(g.Waste),
		moves:    g.Moves,
		score:    g.Score,
		recycles: g.Recycles,
	}
	for i := range g.Foundations {
		s.foundations[i] = cloneCards(g.Foundations[i])
	}
	for i := range g.Tableau {
		s.tableau[i] = cloneCards(g.Tableau[i])
	}
	g.history = append(g.history, s)
}

func cloneCards(c []Card) []Card {
	out := make([]Card, len(c), len(c)+4)
	copy(out, c)
	return out
}

// CardCount は全置き場のカード枚数の合計を返す（整合性チェック用）。
func (g *Game) CardCount() int {
	n := len(g.Stock) + len(g.Waste)
	for _, f := range g.Foundations {
		n += len(f)
	}
	for _, t := range g.Tableau {
		n += len(t)
	}
	return n
}
