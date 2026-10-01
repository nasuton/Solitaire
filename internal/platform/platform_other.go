//go:build !js

// Package platform は実行環境（ブラウザ / デスクトップ）ごとの差異を吸収する。
package platform

// SeedFromURL はデスクトップでは常に false を返す（URL が無いため）。
func SeedFromURL() (int64, bool) { return 0, false }

// SetSeedInURL はデスクトップでは何もしない。
func SetSeedInURL(int64) {}

// IsBrowser はブラウザ上で動作しているかを返す。
func IsBrowser() bool { return false }

// RegisterDebugAPI はデスクトップでは何もしない。
func RegisterDebugAPI(func() map[string]any) {}

// ConfigureCanvas はデスクトップでは何もしない。
func ConfigureCanvas(string) {}
