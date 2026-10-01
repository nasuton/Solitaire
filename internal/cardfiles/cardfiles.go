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

// sourceDirs は元画像（torannpu）側のフォルダ名。
var sourceDirs = map[klondike.Suit]string{
	klondike.Clubs:    "Clubs",
	klondike.Diamonds: "Diamonds",
	klondike.Hearts:   "Hearts",
	klondike.Spades:   "Spades",
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

// SourceRelPath は元画像フォルダ内の相対パス（例: "Clubs/torannpu-Clubs1.png"）を返す。
// スラッシュ区切りなので、OS パスに変換する場合は filepath.FromSlash を使う。
func SourceRelPath(s klondike.Suit, r klondike.Rank) string {
	dir := sourceDirs[s]
	return path.Join(dir, fmt.Sprintf("torannpu-%s%d.png", dir, int(r)))
}

// SourceBackRelPath は元画像フォルダ内の裏面画像の相対パスを返す。
const SourceBackRelPath = "torannpu-BackSide.png"
