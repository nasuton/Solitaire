// Command solitaire はクロンダイク・ソリティア（1 枚めくり）。
// デスクトップと WebAssembly（GitHub Pages）の共通エントリポイント。
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/nasuton/Solitaire/internal/assets"
	"github.com/nasuton/Solitaire/internal/platform"
	"github.com/nasuton/Solitaire/internal/ui"
)

func main() {
	seedFlag := flag.Int64("seed", -1, "配牌の seed（省略時はランダム。ブラウザでは ?seed= でも指定可）")
	flag.Parse()

	seed := *seedFlag
	if seed < 0 {
		if s, ok := platform.SeedFromURL(); ok {
			seed = s
		} else {
			seed = rand.Int64N(1_000_000)
		}
	}

	cards, err := assets.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	game := ui.New(cards, seed)
	platform.SetSeedInURL(seed)

	ebiten.SetWindowTitle("Klondike Solitaire")
	ebiten.SetWindowSize(1280, 800)
	ebiten.SetWindowSizeLimits(480, 360, -1, -1)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
