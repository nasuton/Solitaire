GOROOT := $(shell go env GOROOT)
WASM_OUT := web/solitaire.wasm

.PHONY: wasm serve test desktop assets fontgen clean

## Go を WebAssembly にビルドし、wasm_exec.js を web/ にコピーする。
## -s -w はシンボル/DWARF を削除、-trimpath はビルドを再現可能にする。
wasm:
	GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o $(WASM_OUT) ./cmd/solitaire
	cp "$(GOROOT)/lib/wasm/wasm_exec.js" web/

## wasm をビルドして http://127.0.0.1:8080/ で配信する。
serve: wasm
	go run ./tools/serve -dir web -addr 127.0.0.1:8080

test:
	test -z "$$(gofmt -l .)"
	go vet ./...
	GOOS=js GOARCH=wasm go vet ./cmd/... ./internal/...
	go test ./...

## デスクトップ版をビルドする。
desktop:
	go build -o solitaire$(if $(filter Windows_NT,$(OS)),.exe,) ./cmd/solitaire

## カード画像（52 枚 + 裏面）を自前で生成し直す。
assets:
	go run ./tools/cardgen -out assets/cards -width 180 -height 255

## UI 文言からフォントアトラスを再生成する。
fontgen:
	go run ./tools/fontgen -src internal/ui -out internal/uifont

clean:
	rm -f $(WASM_OUT) web/wasm_exec.js solitaire solitaire.exe
