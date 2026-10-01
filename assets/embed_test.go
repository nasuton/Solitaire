package assets_test

import (
	"bytes"
	"image/png"
	"testing"

	"github.com/nasuton/Solitaire/assets"
	"github.com/nasuton/Solitaire/internal/cardfiles"
	"github.com/nasuton/Solitaire/internal/klondike"
)

func TestAllCardImagesEmbedded(t *testing.T) {
	paths := []string{cardfiles.BackPath()}
	for _, s := range klondike.Suits {
		for r := klondike.Ace; r <= klondike.King; r++ {
			paths = append(paths, cardfiles.Path(s, r))
		}
	}
	if len(paths) != 53 {
		t.Fatalf("paths = %d, want 53", len(paths))
	}
	var w, h int
	for _, p := range paths {
		data, err := assets.FS.ReadFile(p)
		if err != nil {
			t.Fatalf("%s: %v", p, err)
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("%s: %v", p, err)
		}
		if w == 0 {
			w, h = cfg.Width, cfg.Height
			continue
		}
		if cfg.Width != w || cfg.Height != h {
			t.Errorf("%s: size %dx%d differs from %dx%d", p, cfg.Width, cfg.Height, w, h)
		}
	}
	if w == 0 || h == 0 {
		t.Fatal("no image size read")
	}
	// 元画像 712x1008 の縦横比が維持されていること。
	ratio := float64(h) / float64(w)
	if ratio < 1.40 || ratio > 1.43 {
		t.Errorf("aspect ratio %.3f out of expected range", ratio)
	}
}

func TestCardFileNames(t *testing.T) {
	if got := cardfiles.Name(klondike.Clubs, klondike.Ace); got != "clubs_01.png" {
		t.Errorf("Name = %q", got)
	}
	if got := cardfiles.Name(klondike.Spades, klondike.King); got != "spades_13.png" {
		t.Errorf("Name = %q", got)
	}
	if got := cardfiles.SourceRelPath(klondike.Hearts, 10); got != "Hearts/torannpu-Hearts10.png" {
		t.Errorf("SourceRelPath = %q", got)
	}
	if got := cardfiles.Path(klondike.Diamonds, klondike.Queen); got != "cards/diamonds_12.png" {
		t.Errorf("Path = %q", got)
	}
}
