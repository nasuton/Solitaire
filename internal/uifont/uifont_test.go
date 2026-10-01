package uifont

import (
	"path/filepath"
	"testing"

	"github.com/nasuton/Solitaire/internal/fontscan"
)

// TestAtlasCoversUIStrings は internal/ui の文字列リテラルに現れる全文字が
// アトラスに含まれることを確認する。失敗したら `go run ./tools/fontgen` を実行する。
func TestAtlasCoversUIStrings(t *testing.T) {
	runes, err := fontscan.Runes(filepath.Join("..", "ui"))
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range fontscan.ASCII() {
		if !Has(r) {
			t.Errorf("ASCII %q missing from atlas", r)
		}
	}
	for _, r := range runes {
		if !Has(r) {
			t.Errorf("%q (U+%04X) missing from atlas; run `go run ./tools/fontgen`", r, r)
		}
	}
	if !Has('\uFFFD') {
		t.Error("replacement character missing")
	}
}

func TestMeasure(t *testing.T) {
	w, h := Measure("AB", 2)
	if w <= 0 || h != cellH*2 {
		t.Errorf("Measure = %v, %v", w, h)
	}
	wa, _ := Measure("A", 1)
	wj, _ := Measure("あ", 1)
	if !Has('あ') {
		t.Skip("あ not in atlas")
	}
	if wj <= wa {
		t.Errorf("fullwidth %v should be wider than halfwidth %v", wj, wa)
	}
	if w0, _ := Measure("", 1); w0 != 0 {
		t.Errorf("empty width = %v", w0)
	}
	if g := lookup('\U0010FFFF'); g != lookup('\uFFFD') {
		t.Error("unknown rune should fall back to replacement glyph")
	}
}
