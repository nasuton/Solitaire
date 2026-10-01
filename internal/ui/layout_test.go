package ui

import (
	"testing"

	"github.com/nasuton/Solitaire/internal/klondike"
)

func TestRectOverlapArea(t *testing.T) {
	a := Rect{0, 0, 10, 10}
	if got := a.OverlapArea(Rect{5, 5, 10, 10}); got != 25 {
		t.Errorf("overlap = %v, want 25", got)
	}
	if got := a.OverlapArea(Rect{10, 0, 5, 5}); got != 0 {
		t.Errorf("touching rects should not overlap, got %v", got)
	}
	if !a.Contains(0, 0) || a.Contains(10, 10) {
		t.Error("Contains should be half-open")
	}
}

func TestNewLayoutFitsScreen(t *testing.T) {
	for _, size := range [][2]float64{{1280, 800}, {800, 1200}, {1280, 560}, {400, 300}} {
		l := NewLayout(size[0], size[1])
		if l.CardW < 24 || l.CardH <= l.CardW {
			t.Errorf("%v: bad card size %vx%v", size, l.CardW, l.CardH)
		}
		last := l.Tableau[6]
		if last.X+last.W > l.W+0.5 {
			t.Errorf("%v: tableau overflows width: %v > %v", size, last.X+last.W, l.W)
		}
		if l.Stock.Y < l.HeaderH {
			t.Errorf("%v: stock overlaps header", size)
		}
		for i := 0; i < 4; i++ {
			if l.PileRect(klondike.FoundationID(i)) != l.Foundations[i] {
				t.Errorf("PileRect foundation %d mismatch", i)
			}
		}
	}
	if l := NewLayout(1280, 800); l.TwoRowHeader {
		t.Error("wide layout should use single-row header")
	}
	if l := NewLayout(800, 1200); !l.TwoRowHeader {
		t.Error("narrow layout should use two-row header")
	}
}

func TestTableauOffsetsCompress(t *testing.T) {
	l := NewLayout(1280, 560)
	short := []klondike.Card{{FaceUp: false}, {FaceUp: true}}
	d, u := l.TableauOffsets(short)
	if d != l.DownOffset || u != l.UpOffset {
		t.Errorf("short pile should use default offsets, got %v/%v", d, u)
	}

	long := make([]klondike.Card, 19)
	for i := range long {
		long[i].FaceUp = i >= 6
	}
	d2, u2 := l.TableauOffsets(long)
	if u2 >= l.UpOffset {
		t.Errorf("long pile should compress up offset: %v >= %v", u2, l.UpOffset)
	}
	if d2 < 2 || u2 < 6 {
		t.Errorf("offsets below minimum: %v/%v", d2, u2)
	}
	id := klondike.TableauID(0)
	top := l.TopRect(id, long)
	if top.Y+top.H > l.Bottom+1 {
		t.Errorf("compressed pile still overflows: %v > %v", top.Y+top.H, l.Bottom)
	}
	drop := l.DropRect(id, long)
	if drop.Y != l.Tableau[0].Y || drop.Y+drop.H != top.Y+top.H {
		t.Errorf("DropRect should span whole column: %+v (top %+v)", drop, top)
	}
	next := l.NextRect(id, long)
	if next.Y != top.Y+u2 {
		t.Errorf("NextRect.Y = %v, want %v", next.Y, top.Y+u2)
	}
	if l.NextRect(klondike.FoundationID(1), long) != l.Foundations[1] {
		t.Error("NextRect on foundation should be the pile rect")
	}
	if l.TopRect(id, nil) != l.Tableau[0] {
		t.Error("TopRect of empty pile should be the pile rect")
	}
}
