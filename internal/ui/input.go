package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// pointer はマウス左ボタンと最初のタッチを 1 つのポインタとして扱う。
type pointer struct {
	X, Y         float64
	Pressed      bool
	JustPressed  bool
	JustReleased bool

	touching bool
	touchID  ebiten.TouchID
	touchBuf []ebiten.TouchID
}

func (p *pointer) update() {
	p.JustPressed = false
	p.JustReleased = false

	if p.touching {
		if inpututil.IsTouchJustReleased(p.touchID) {
			x, y := inpututil.TouchPositionInPreviousTick(p.touchID)
			p.X, p.Y = float64(x), float64(y)
			p.Pressed = false
			p.JustReleased = true
			p.touching = false
			return
		}
		x, y := ebiten.TouchPosition(p.touchID)
		p.X, p.Y = float64(x), float64(y)
		p.Pressed = true
		return
	}

	p.touchBuf = inpututil.AppendJustPressedTouchIDs(p.touchBuf[:0])
	if len(p.touchBuf) > 0 && !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		p.touching = true
		p.touchID = p.touchBuf[0]
		x, y := ebiten.TouchPosition(p.touchID)
		p.X, p.Y = float64(x), float64(y)
		p.Pressed = true
		p.JustPressed = true
		return
	}

	mx, my := ebiten.CursorPosition()
	p.X, p.Y = float64(mx), float64(my)
	p.Pressed = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	p.JustPressed = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	p.JustReleased = inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
}
