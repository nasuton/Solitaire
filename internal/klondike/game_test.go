package klondike

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewDeck(t *testing.T) {
	deck := NewDeck()
	if len(deck) != 52 {
		t.Fatalf("len = %d, want 52", len(deck))
	}
	seen := map[int]bool{}
	for _, c := range deck {
		if c.FaceUp {
			t.Errorf("%v should be face down", c)
		}
		if seen[c.ID()] {
			t.Errorf("duplicate id %d", c.ID())
		}
		seen[c.ID()] = true
		if c.ID() < 0 || c.ID() > 51 {
			t.Errorf("id out of range: %d", c.ID())
		}
	}
}

func TestCardStrings(t *testing.T) {
	cases := []struct {
		c    Card
		want string
	}{
		{Card{Spades, Ace, true}, "A♠"},
		{Card{Hearts, 10, true}, "10♥"},
		{Card{Diamonds, Queen, true}, "Q♦"},
		{Card{Clubs, King, false}, "(K♣)"},
		{Card{Clubs, Jack, true}, "J♣"},
	}
	for _, tc := range cases {
		if got := tc.c.String(); got != tc.want {
			t.Errorf("%+v.String() = %q, want %q", tc.c, got, tc.want)
		}
	}
	if Clubs.Color() != Black || Spades.Color() != Black || Hearts.Color() != Red || Diamonds.Color() != Red {
		t.Error("suit colors wrong")
	}
	for _, s := range Suits {
		if strings.HasPrefix(s.String(), "Suit(") {
			t.Errorf("unexpected suit name %q", s)
		}
	}
	if Suit(9).String() != "Suit(9)" || PileKind(9).String() != "PileKind(9)" {
		t.Error("fallback names wrong")
	}
	if TableauID(3).String() != "Tableau[3]" || FoundationID(1).String() != "Foundation[1]" || StockID.String() != "Stock" || WasteID.String() != "Waste" {
		t.Error("pile id strings wrong")
	}
}

func TestDeal(t *testing.T) {
	g := New(1)
	if g.CardCount() != 52 {
		t.Fatalf("card count = %d", g.CardCount())
	}
	if len(g.Stock) != 24 {
		t.Errorf("stock = %d, want 24", len(g.Stock))
	}
	if len(g.Waste) != 0 {
		t.Errorf("waste = %d, want 0", len(g.Waste))
	}
	for i, col := range g.Tableau {
		if len(col) != i+1 {
			t.Errorf("tableau[%d] len = %d, want %d", i, len(col), i+1)
		}
		for j, c := range col {
			wantUp := j == i
			if c.FaceUp != wantUp {
				t.Errorf("tableau[%d][%d] faceUp = %v, want %v", i, j, c.FaceUp, wantUp)
			}
		}
	}
	for _, c := range g.Stock {
		if c.FaceUp {
			t.Errorf("stock card %v face up", c)
		}
	}
	if g.Moves != 0 || g.Score != 0 || g.CanUndo() {
		t.Error("fresh game should have zero moves/score and no undo")
	}
}

func TestDeterministicSeed(t *testing.T) {
	a, b := New(12345), New(12345)
	if fmt.Sprint(a.Tableau, a.Stock) != fmt.Sprint(b.Tableau, b.Stock) {
		t.Error("same seed produced different deals")
	}
	c := New(12346)
	if fmt.Sprint(a.Tableau, a.Stock) == fmt.Sprint(c.Tableau, c.Stock) {
		t.Error("different seeds produced the same deal")
	}
	// 既知 seed の配牌を固定し、乱数アルゴリズム変更による非互換を検知する。
	const wantTops = "3♠ A♣ 7♦ 6♠ 8♠ 7♠ 3♥"
	var tops []string
	for i := range a.Tableau {
		top, _ := a.Top(TableauID(i))
		tops = append(tops, top.String())
	}
	if got := strings.Join(tops, " "); got != wantTops {
		t.Errorf("seed 12345 tableau tops = %q, want %q", got, wantTops)
	}
}

func TestRestartSameDeal(t *testing.T) {
	g := New(777)
	want := fmt.Sprint(g.Tableau, g.Stock)
	_ = g.Draw()
	_ = g.Draw()
	g.Restart()
	if got := fmt.Sprint(g.Tableau, g.Stock); got != want {
		t.Error("restart did not restore the original deal")
	}
	if g.Moves != 0 || g.Score != 0 || g.CanUndo() || len(g.Waste) != 0 {
		t.Error("restart did not reset counters")
	}
}

func TestDrawAndRecycle(t *testing.T) {
	g := New(2)
	stockBefore := cloneCards(g.Stock)
	if !g.CanDraw() || g.WillRecycle() {
		t.Fatal("fresh game should allow draw without recycling")
	}
	if err := g.Draw(); err != nil {
		t.Fatal(err)
	}
	if len(g.Stock) != 23 || len(g.Waste) != 1 || !g.Waste[0].FaceUp || g.Moves != 1 {
		t.Fatalf("after draw: stock=%d waste=%d moves=%d", len(g.Stock), len(g.Waste), g.Moves)
	}
	if g.Waste[0].ID() != stockBefore[23].ID() {
		t.Error("drawn card should be the top of the stock")
	}
	for len(g.Stock) > 0 {
		if err := g.Draw(); err != nil {
			t.Fatal(err)
		}
	}
	if len(g.Waste) != 24 || !g.WillRecycle() {
		t.Fatalf("waste=%d, willRecycle=%v", len(g.Waste), g.WillRecycle())
	}
	g.Score = 150
	if err := g.Draw(); err != nil {
		t.Fatal(err)
	}
	if len(g.Stock) != 24 || len(g.Waste) != 0 || g.Recycles != 1 {
		t.Fatalf("after recycle: stock=%d waste=%d recycles=%d", len(g.Stock), len(g.Waste), g.Recycles)
	}
	if g.Score != 50 {
		t.Errorf("score after recycle = %d, want 50", g.Score)
	}
	for i, c := range g.Stock {
		if c.FaceUp {
			t.Errorf("recycled stock card %d face up", i)
		}
		if c.ID() != stockBefore[i].ID() {
			t.Errorf("recycled stock order differs at %d: %v vs %v", i, c, stockBefore[i])
		}
	}
	// 再循環のスコアは 0 を下回らない。
	for len(g.Stock) > 0 {
		_ = g.Draw()
	}
	_ = g.Draw()
	if g.Score != 0 {
		t.Errorf("score should clamp at 0, got %d", g.Score)
	}
	// 山札・捨て札が空なら何もできない。
	e := &Game{}
	if e.CanDraw() || e.Draw() != ErrNothingToDo {
		t.Error("empty game should not draw")
	}
}

// setup はテスト用に任意の状態を作る。
func setup(fn func(g *Game)) *Game {
	g := &Game{Seed: 0}
	for i := range g.Foundations {
		g.Foundations[i] = []Card{}
	}
	for i := range g.Tableau {
		g.Tableau[i] = []Card{}
	}
	fn(g)
	return g
}

func up(s Suit, r Rank) Card   { return Card{Suit: s, Rank: r, FaceUp: true} }
func down(s Suit, r Rank) Card { return Card{Suit: s, Rank: r, FaceUp: false} }

func TestMoveValidation(t *testing.T) {
	g := setup(func(g *Game) {
		g.Waste = []Card{up(Hearts, 5)}
		g.Tableau[0] = []Card{down(Clubs, 9), up(Spades, 6)}
		g.Tableau[1] = []Card{up(Diamonds, 6)}
		g.Tableau[2] = []Card{up(Hearts, King), up(Clubs, Queen), up(Diamonds, Jack)}
		g.Tableau[3] = []Card{}
		g.Tableau[4] = []Card{down(Spades, 2)}
		g.Tableau[5] = []Card{up(Spades, 10), up(Hearts, 2)} // 不正な連なり（構築上ありえないが検証する）
		g.Tableau[6] = []Card{up(Clubs, 5)}
		g.Foundations[0] = []Card{up(Hearts, Ace), up(Hearts, 2), up(Hearts, 3), up(Hearts, 4)}
		g.Stock = []Card{down(Clubs, Ace)}
	})

	cases := []struct {
		name string
		m    Move
		ok   bool
	}{
		{"waste onto opposite color descending", Move{WasteID, TableauID(0), 1}, true},
		{"waste onto same color", Move{WasteID, TableauID(1), 1}, false},
		{"waste to matching foundation", Move{WasteID, FoundationID(0), 1}, true},
		{"waste to empty foundation (not ace)", Move{WasteID, FoundationID(1), 1}, false},
		{"waste to empty tableau (not king)", Move{WasteID, TableauID(3), 1}, false},
		{"king run to empty tableau", Move{TableauID(2), TableauID(3), 3}, true},
		{"partial run Q-J onto nothing valid", Move{TableauID(2), TableauID(0), 2}, false},
		{"J onto black Q? no, J♦ onto 6♠ invalid", Move{TableauID(2), TableauID(0), 1}, false},
		{"face-down card cannot move", Move{TableauID(4), TableauID(3), 1}, false},
		{"run including face-down card", Move{TableauID(0), TableauID(3), 2}, false},
		{"count too large", Move{TableauID(1), TableauID(0), 2}, false},
		{"count zero", Move{TableauID(1), TableauID(0), 0}, false},
		{"from stock", Move{StockID, TableauID(3), 1}, false},
		{"to stock", Move{WasteID, StockID, 1}, false},
		{"to waste", Move{TableauID(1), WasteID, 1}, false},
		{"same pile", Move{TableauID(1), TableauID(1), 1}, false},
		{"invalid pile index", Move{TableauID(9), TableauID(0), 1}, false},
		{"invalid dest index", Move{WasteID, FoundationID(4), 1}, false},
		{"multi-card to foundation", Move{TableauID(2), FoundationID(1), 2}, false},
		{"invalid run in tableau", Move{TableauID(5), TableauID(3), 2}, false},
		{"foundation back to tableau", Move{FoundationID(0), TableauID(6), 1}, true},
		{"foundation onto wrong rank", Move{FoundationID(0), TableauID(0), 1}, false},
		{"foundation multi", Move{FoundationID(0), TableauID(6), 2}, false},
		{"onto face-down tableau top", Move{WasteID, TableauID(4), 1}, false},
	}
	for _, tc := range cases {
		if got := g.CanMove(tc.m); got != tc.ok {
			t.Errorf("%s: CanMove(%v) = %v, want %v", tc.name, tc.m, got, tc.ok)
		}
	}
	if err := g.Apply(Move{WasteID, TableauID(1), 1}); err != ErrInvalidMove {
		t.Errorf("Apply invalid = %v, want ErrInvalidMove", err)
	}
	if err := g.Apply(Move{TableauID(9), TableauID(0), 1}); err != ErrInvalidPile {
		t.Errorf("Apply invalid pile = %v, want ErrInvalidPile", err)
	}
	if g.Moves != 0 || g.CanUndo() {
		t.Error("failed apply must not change state")
	}
}

func TestApplyScoringAndFlip(t *testing.T) {
	g := setup(func(g *Game) {
		g.Waste = []Card{up(Hearts, 5), up(Clubs, Ace)}
		g.Tableau[0] = []Card{down(Clubs, 9), up(Spades, 6)}
		g.Tableau[1] = []Card{down(Diamonds, 3), up(Hearts, 7)}
		g.Foundations[0] = []Card{up(Hearts, Ace), up(Hearts, 2), up(Hearts, 3), up(Hearts, 4)}
	})

	// 捨て札 → 組札 +10
	if err := g.Apply(Move{WasteID, FoundationID(1), 1}); err != nil {
		t.Fatal(err)
	}
	if g.Score != 10 || g.Moves != 1 {
		t.Errorf("after waste→foundation: score=%d moves=%d", g.Score, g.Moves)
	}
	// 捨て札 → 場札 +5
	if err := g.Apply(Move{WasteID, TableauID(0), 1}); err != nil {
		t.Fatal(err)
	}
	if g.Score != 15 {
		t.Errorf("after waste→tableau: score=%d", g.Score)
	}
	// 場札 → 場札（6♠ 5♥ を 7♥ の上へ）: 0 点だがめくり +5
	if err := g.Apply(Move{TableauID(0), TableauID(1), 2}); err != nil {
		t.Fatal(err)
	}
	if g.Score != 20 {
		t.Errorf("after tableau→tableau with flip: score=%d", g.Score)
	}
	if !g.Tableau[0][0].FaceUp {
		t.Error("exposed tableau card should be flipped face up")
	}
	if len(g.Tableau[1]) != 4 {
		t.Errorf("tableau[1] len = %d, want 4", len(g.Tableau[1]))
	}
	// 組札 → 場札 −15
	g.Tableau[2] = []Card{up(Spades, 5)}
	if err := g.Apply(Move{FoundationID(0), TableauID(2), 1}); err != nil {
		t.Fatal(err)
	}
	if g.Score != 5 {
		t.Errorf("after foundation→tableau: score=%d", g.Score)
	}
	// 場札 → 組札 +10
	if err := g.Apply(Move{TableauID(2), FoundationID(0), 1}); err != nil {
		t.Fatal(err)
	}
	if g.Score != 15 {
		t.Errorf("after tableau→foundation: score=%d", g.Score)
	}
	// 下限 0
	g.Score = 10
	g.Tableau[2] = []Card{up(Spades, 5)}
	if err := g.Apply(Move{FoundationID(0), TableauID(2), 1}); err != nil {
		t.Fatal(err)
	}
	if g.Score != 0 {
		t.Errorf("score should clamp at 0, got %d", g.Score)
	}
	if g.CardCount() != 11 {
		t.Errorf("card count changed: %d", g.CardCount())
	}
}

func TestUndoRoundTrip(t *testing.T) {
	g := New(99)
	snapshots := []string{fmt.Sprint(g.Tableau, g.Stock, g.Waste, g.Foundations, g.Moves, g.Score)}

	// いくつか手を進める（Draw と可能なら自動移動）。
	for i := 0; i < 30; i++ {
		moved := false
		for col := 0; col < 7 && !moved; col++ {
			if m, ok := g.FoundationMoveFor(TableauID(col)); ok {
				if err := g.Apply(m); err != nil {
					t.Fatal(err)
				}
				moved = true
			}
		}
		if !moved {
			if m, ok := g.FoundationMoveFor(WasteID); ok {
				if err := g.Apply(m); err != nil {
					t.Fatal(err)
				}
			} else if err := g.Draw(); err != nil {
				t.Fatal(err)
			}
		}
		snapshots = append(snapshots, fmt.Sprint(g.Tableau, g.Stock, g.Waste, g.Foundations, g.Moves, g.Score))
	}
	if g.Moves != 30 {
		t.Fatalf("moves = %d, want 30", g.Moves)
	}
	for i := len(snapshots) - 2; i >= 0; i-- {
		if err := g.Undo(); err != nil {
			t.Fatal(err)
		}
		if got := fmt.Sprint(g.Tableau, g.Stock, g.Waste, g.Foundations, g.Moves, g.Score); got != snapshots[i] {
			t.Fatalf("undo to step %d mismatch", i)
		}
	}
	if g.CanUndo() {
		t.Error("history should be empty")
	}
	if err := g.Undo(); err != ErrNothingToUndo {
		t.Errorf("Undo on empty = %v, want ErrNothingToUndo", err)
	}
}

func TestUndoIsolation(t *testing.T) {
	// Undo で復元したスライスを変更しても過去の履歴に影響しない。
	g := New(5)
	_ = g.Draw()
	_ = g.Draw()
	_ = g.Undo()
	g.Waste[0].FaceUp = false
	_ = g.Undo()
	if len(g.Waste) != 0 {
		t.Error("undo should restore empty waste")
	}
}

func TestFoundationMoveFor(t *testing.T) {
	g := setup(func(g *Game) {
		g.Waste = []Card{up(Clubs, Ace)}
		g.Tableau[0] = []Card{up(Spades, 2)}
		g.Tableau[1] = []Card{down(Spades, 3)}
		g.Foundations[2] = []Card{up(Spades, Ace)}
	})
	m, ok := g.FoundationMoveFor(WasteID)
	if !ok || m.To != FoundationID(0) || m.Count != 1 {
		t.Errorf("ace from waste: %v %v", m, ok)
	}
	m, ok = g.FoundationMoveFor(TableauID(0))
	if !ok || m.To != FoundationID(2) {
		t.Errorf("2♠ should go to spade foundation: %v %v", m, ok)
	}
	if _, ok := g.FoundationMoveFor(TableauID(1)); ok {
		t.Error("face-down card must not auto-move")
	}
	if _, ok := g.FoundationMoveFor(TableauID(3)); ok {
		t.Error("empty pile must not auto-move")
	}
	if _, ok := g.FoundationMoveFor(StockID); ok {
		t.Error("stock must not auto-move")
	}
	if _, ok := g.FoundationMoveFor(FoundationID(2)); ok {
		t.Error("foundation must not auto-move")
	}
}

func TestWinAndAutoComplete(t *testing.T) {
	g := setup(func(g *Game) {
		for _, s := range Suits {
			for r := Ace; r <= 10; r++ {
				g.Foundations[s] = append(g.Foundations[s], up(s, r))
			}
		}
		g.Tableau[0] = []Card{up(Spades, King), up(Hearts, Queen), up(Spades, Jack)}
		g.Tableau[1] = []Card{up(Hearts, King), up(Spades, Queen), up(Hearts, Jack)}
		g.Tableau[2] = []Card{up(Clubs, King), up(Diamonds, Queen), up(Clubs, Jack)}
		g.Tableau[3] = []Card{up(Diamonds, King), up(Clubs, Queen), up(Diamonds, Jack)}
	})
	if g.CardCount() != 52 {
		t.Fatalf("card count = %d", g.CardCount())
	}
	if g.IsWon() {
		t.Fatal("should not be won yet")
	}
	if !g.CanAutoComplete() {
		t.Fatal("should be able to auto-complete")
	}
	steps := 0
	for {
		m, ok := g.NextAutoCompleteMove()
		if !ok {
			break
		}
		if m.To.Kind != Foundation {
			t.Fatalf("auto move should target foundation: %v", m)
		}
		if err := g.Apply(m); err != nil {
			t.Fatal(err)
		}
		steps++
		if steps > 12 {
			t.Fatal("too many steps")
		}
	}
	if steps != 12 {
		t.Errorf("steps = %d, want 12", steps)
	}
	if !g.IsWon() {
		t.Error("should be won")
	}
	if g.CanAutoComplete() {
		t.Error("won game must not auto-complete")
	}
	if g.Score != 120 {
		t.Errorf("score = %d, want 120", g.Score)
	}

	// 山札が残っていると自動完了不可。
	g2 := New(3)
	if g2.CanAutoComplete() {
		t.Error("fresh deal must not auto-complete (stock not empty)")
	}
	if _, ok := g2.NextAutoCompleteMove(); ok {
		t.Error("fresh deal must not have auto-complete move")
	}
	// 山札・捨て札が空でも裏向きカードがあれば不可。
	g3 := setup(func(g *Game) {
		g.Tableau[0] = []Card{down(Spades, King), up(Hearts, Queen)}
	})
	if g3.CanAutoComplete() {
		t.Error("face-down cards must block auto-complete")
	}
}

func TestPileAccessors(t *testing.T) {
	g := New(8)
	if g.Pile(PileID{Kind: PileKind(42)}) != nil {
		t.Error("unknown kind should return nil")
	}
	if g.Pile(FoundationID(-1)) != nil || g.Pile(TableauID(7)) != nil {
		t.Error("out-of-range index should return nil")
	}
	if _, ok := g.Top(WasteID); ok {
		t.Error("empty waste has no top")
	}
	if c, ok := g.Top(TableauID(6)); !ok || !c.FaceUp {
		t.Error("tableau[6] top should be face up")
	}
	if len(g.Pile(StockID)) != 24 {
		t.Error("Pile(StockID) should be the stock")
	}
	if (PileID{Kind: PileKind(42)}).Valid() {
		t.Error("unknown kind should be invalid")
	}
	if (PileID{Kind: PileKind(42)}).String() != "PileKind(42)" {
		t.Error("unknown pile string")
	}
}
