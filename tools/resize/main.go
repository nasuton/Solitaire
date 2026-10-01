// Command resize は元のカード画像（712x1008px）を縮小して assets/cards に配置する。
//
//	go run ./tools/resize -src D:\work\Go\torannpu -dst assets/cards -width 180
//
// 縮小には golang.org/x/image/draw の CatmullRom を使い、縦横比は維持する。
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	xdraw "golang.org/x/image/draw"

	"github.com/nasuton/Solitaire/internal/cardfiles"
	"github.com/nasuton/Solitaire/internal/klondike"
)

func main() {
	src := flag.String("src", "", "元画像フォルダ（Clubs/ Diamonds/ Hearts/ Spades/ と torannpu-BackSide.png を含む）")
	dst := flag.String("dst", filepath.Join("assets", "cards"), "出力先フォルダ")
	width := flag.Int("width", 180, "出力画像の幅（px）。高さは縦横比を維持して決まる")
	flag.Parse()

	if *src == "" {
		fmt.Fprintln(os.Stderr, "-src を指定してください")
		flag.Usage()
		os.Exit(2)
	}
	if err := run(*src, *dst, *width); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

type job struct {
	src string
	dst string
}

func run(srcDir, dstDir string, width int) error {
	if width <= 0 {
		return fmt.Errorf("width は正の値にしてください: %d", width)
	}
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return err
	}

	var jobs []job
	for _, s := range klondike.Suits {
		for r := klondike.Ace; r <= klondike.King; r++ {
			jobs = append(jobs, job{
				src: filepath.Join(srcDir, filepath.FromSlash(cardfiles.SourceRelPath(s, r))),
				dst: filepath.Join(dstDir, cardfiles.Name(s, r)),
			})
		}
	}
	jobs = append(jobs, job{
		src: filepath.Join(srcDir, cardfiles.SourceBackRelPath),
		dst: filepath.Join(dstDir, cardfiles.BackName),
	})

	var total int64
	for _, j := range jobs {
		n, err := resizeFile(j.src, j.dst, width)
		if err != nil {
			return fmt.Errorf("%s: %w", j.src, err)
		}
		total += n
	}
	fmt.Printf("%d files written to %s (total %d bytes, %.1f KiB)\n", len(jobs), dstDir, total, float64(total)/1024)
	return nil
}

func resizeFile(srcPath, dstPath string, width int) (int64, error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return 0, err
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		return 0, err
	}

	b := img.Bounds()
	height := int(float64(width)*float64(b.Dy())/float64(b.Dx()) + 0.5)
	out := image.NewNRGBA(image.Rect(0, 0, width, height))
	xdraw.CatmullRom.Scale(out, out.Bounds(), img, b, xdraw.Over, nil)

	w, err := os.Create(dstPath)
	if err != nil {
		return 0, err
	}
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(w, out); err != nil {
		w.Close()
		return 0, err
	}
	if err := w.Close(); err != nil {
		return 0, err
	}
	st, err := os.Stat(dstPath)
	if err != nil {
		return 0, err
	}
	return st.Size(), nil
}
