// Package assets はカード画像を embed でバイナリに埋め込む。
// 画像は tools/cardgen で生成した自作のもの（cards/*.png）。
package assets

import "embed"

// FS は埋め込んだアセットのファイルシステム。パスは "cards/clubs_01.png" のような形式。
//
//go:embed cards/*.png
var FS embed.FS
