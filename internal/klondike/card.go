// Package klondike はクロンダイク・ソリティアの純粋なルールロジックを提供する。
// Ebitengine には依存せず、単体テストのみで検証できる。
package klondike

import "fmt"

// Suit はカードのスート。
type Suit int

// 4 種類のスート。並び順はアセットのファイル名と一致させている。
const (
	Clubs Suit = iota
	Diamonds
	Hearts
	Spades
)

// Suits は全スートを列挙する。
var Suits = [4]Suit{Clubs, Diamonds, Hearts, Spades}

// Color はスートの色。
type Color int

// 黒・赤。
const (
	Black Color = iota
	Red
)

// Color はスートの色を返す。
func (s Suit) Color() Color {
	if s == Diamonds || s == Hearts {
		return Red
	}
	return Black
}

// String はスートの英語名を返す。
func (s Suit) String() string {
	switch s {
	case Clubs:
		return "Clubs"
	case Diamonds:
		return "Diamonds"
	case Hearts:
		return "Hearts"
	case Spades:
		return "Spades"
	}
	return fmt.Sprintf("Suit(%d)", int(s))
}

// Rank はカードの数字。1=A, 11=J, 12=Q, 13=K。
type Rank int

// 特別なランク。
const (
	Ace   Rank = 1
	Jack  Rank = 11
	Queen Rank = 12
	King  Rank = 13
)

// String はランクの表記を返す。
func (r Rank) String() string {
	switch r {
	case Ace:
		return "A"
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	}
	return fmt.Sprintf("%d", int(r))
}

// Card は 1 枚のカード。FaceUp は表向きかどうか。
type Card struct {
	Suit   Suit
	Rank   Rank
	FaceUp bool
}

// ID はカードを一意に識別する 0..51 の番号を返す（表裏は無視）。
func (c Card) ID() int {
	return int(c.Suit)*13 + int(c.Rank) - 1
}

// String はデバッグ用の短い表記（例: "A♠", "10♥"）を返す。裏向きは括弧で囲む。
func (c Card) String() string {
	sym := [...]string{"♣", "♦", "♥", "♠"}
	s := c.Rank.String() + sym[c.Suit]
	if !c.FaceUp {
		s = "(" + s + ")"
	}
	return s
}

// NewDeck は 52 枚の裏向きカードをスート順・ランク順で返す。
func NewDeck() []Card {
	deck := make([]Card, 0, 52)
	for _, s := range Suits {
		for r := Ace; r <= King; r++ {
			deck = append(deck, Card{Suit: s, Rank: r})
		}
	}
	return deck
}
