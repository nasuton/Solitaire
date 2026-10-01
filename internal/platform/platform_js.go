//go:build js

// Package platform は実行環境（ブラウザ / デスクトップ）ごとの差異を吸収する。
package platform

import (
	"strconv"
	"syscall/js"
)

// SeedFromURL はブラウザの URL クエリ `?seed=` から seed を読み取る。
func SeedFromURL() (int64, bool) {
	loc := js.Global().Get("window").Get("location")
	if !loc.Truthy() {
		return 0, false
	}
	params := js.Global().Get("URLSearchParams").New(loc.Get("search"))
	v := params.Call("get", "seed")
	if !v.Truthy() {
		return 0, false
	}
	n, err := strconv.ParseInt(v.String(), 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

// SetSeedInURL は履歴を増やさずに URL クエリの seed を書き換える。
// iframe 内で動作している場合は同一オリジンの親ページの URL も更新する。
func SetSeedInURL(seed int64) {
	window := js.Global().Get("window")
	replaceSeed(window, seed)
	parent := window.Get("parent")
	if parent.Truthy() && !parent.Equal(window) {
		replaceSeed(parent, seed)
	}
}

func replaceSeed(window js.Value, seed int64) {
	defer func() { _ = recover() }() // クロスオリジンの親はアクセス不可なので無視
	loc := window.Get("location")
	history := window.Get("history")
	if !loc.Truthy() || !history.Truthy() {
		return
	}
	url := js.Global().Get("URL").New(loc.Get("href"))
	url.Get("searchParams").Call("set", "seed", strconv.FormatInt(seed, 10))
	history.Call("replaceState", js.Null(), "", url.Call("toString"))
}

// IsBrowser はブラウザ上で動作しているかを返す。
func IsBrowser() bool { return true }

// ConfigureCanvas は Ebitengine が生成した canvas にアクセシビリティ属性を付与し、
// 読み込み中表示を取り除く。初回フレームで一度だけ呼ぶ。
func ConfigureCanvas(label string) {
	doc := js.Global().Get("document")
	if !doc.Truthy() {
		return
	}
	canvas := doc.Call("querySelector", "canvas")
	if canvas.Truthy() {
		canvas.Call("setAttribute", "role", "img")
		canvas.Call("setAttribute", "aria-label", label)
		canvas.Call("setAttribute", "id", "game-canvas")
	}
	if loading := doc.Call("getElementById", "loading"); loading.Truthy() {
		loading.Call("remove")
	}
}

// RegisterDebugAPI は window.solitaireState() として状態取得関数を公開する（E2E 検証用）。
func RegisterDebugAPI(state func() map[string]any) {
	js.Global().Set("solitaireState", js.FuncOf(func(this js.Value, args []js.Value) any {
		return js.ValueOf(state())
	}))
}
