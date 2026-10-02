// Package cardfiles はカード画像のファイル名規則を 1 箇所にまとめる。
// 画像を差し替える場合はこのファイルだけを変更すればよい。
package cardfiles

import (
	"fmt"
	"path"

	"github.com/nasuton/Solitaire/internal/klondike"
)

// Dir は assets パッケージ内でカード画像を置くディレクトリ名。
const Dir = "cards"

// BackName は裏面画像のファイル名。
const BackName = "back.png"

// suitNames はリポジトリ内のファイル名に使うスート名（小文字）。
var suitNames = map[klondike.Suit]string{
	klondike.Clubs:    "clubs",
	klondike.Diamonds: "diamonds",
	klondike.Hearts:   "hearts",
	klondike.Spades:   "spades",
}

// Name はスートとランクからリポジトリ内のファイル名（例: "clubs_01.png"）を返す。
func Name(s klondike.Suit, r klondike.Rank) string {
	return fmt.Sprintf("%s_%02d.png", suitNames[s], int(r))
}

// Path は embed.FS 上のパス（例: "cards/clubs_01.png"）を返す。
func Path(s klondike.Suit, r klondike.Rank) string {
	return path.Join(Dir, Name(s, r))
}

// BackPath は embed.FS 上の裏面画像パスを返す。
func BackPath() string {
	return path.Join(Dir, BackName)
}
