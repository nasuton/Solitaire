package cardfiles

import (
	"testing"

	"github.com/nasuton/Solitaire/internal/klondike"
)

func TestNames(t *testing.T) {
	cases := []struct {
		suit klondike.Suit
		rank klondike.Rank
		name string
		src  string
	}{
		{klondike.Clubs, 1, "clubs_01.png", "Clubs/torannpu-Clubs1.png"},
		{klondike.Diamonds, 10, "diamonds_10.png", "Diamonds/torannpu-Diamonds10.png"},
		{klondike.Hearts, 12, "hearts_12.png", "Hearts/torannpu-Hearts12.png"},
		{klondike.Spades, 13, "spades_13.png", "Spades/torannpu-Spades13.png"},
	}
	for _, c := range cases {
		if got := Name(c.suit, c.rank); got != c.name {
			t.Errorf("Name(%v,%v) = %q, want %q", c.suit, c.rank, got, c.name)
		}
		if got := Path(c.suit, c.rank); got != "cards/"+c.name {
			t.Errorf("Path = %q", got)
		}
		if got := SourceRelPath(c.suit, c.rank); got != c.src {
			t.Errorf("SourceRelPath = %q, want %q", got, c.src)
		}
	}
	if BackPath() != "cards/back.png" {
		t.Errorf("BackPath = %q", BackPath())
	}
}

func TestNamesUnique(t *testing.T) {
	seen := map[string]bool{BackPath(): true}
	for _, s := range klondike.Suits {
		for r := klondike.Rank(1); r <= 13; r++ {
			p := Path(s, r)
			if seen[p] {
				t.Fatalf("duplicate path %q", p)
			}
			seen[p] = true
		}
	}
	if len(seen) != 53 {
		t.Errorf("expected 53 distinct paths, got %d", len(seen))
	}
}
