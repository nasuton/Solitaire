// Package fontscan は Go ソース中の文字列リテラルから使用文字を収集する。
// tools/fontgen（アトラス生成）と internal/uifont のテスト（漏れ検知）で共用する。
package fontscan

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ASCII は常にアトラスに含める基本文字（印字可能 ASCII）。
func ASCII() []rune {
	rs := make([]rune, 0, 95)
	for r := rune(0x20); r <= 0x7e; r++ {
		rs = append(rs, r)
	}
	return rs
}

// Runes は dir 直下の *.go（_test.go を除く）の文字列リテラルに現れる、
// ASCII 以外の文字をソートして返す。
func Runes(dir string) ([]rune, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	set := map[rune]bool{}
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			return nil, err
		}
		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			s, err := strconv.Unquote(lit.Value)
			if err != nil {
				return true
			}
			for _, r := range s {
				if r > 0x7e {
					set[r] = true
				}
			}
			return true
		})
	}
	out := make([]rune, 0, len(set))
	for r := range set {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}
