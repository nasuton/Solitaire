# Solitaire

Go と [Ebitengine](https://ebitengine.org/) で実装したクロンダイク・ソリティア（山札 1 枚めくり）です。
WebAssembly にビルドして GitHub Pages で公開でき、Windows / macOS / Linux のデスクトップ版としても動作します。

- 🎮 プレイ: <https://nasuton.github.io/Solitaire/>
- 🛠 技術: Go 1.26 / Ebitengine v2.10 / WebAssembly / GitHub Actions（Pages 自動デプロイ）

## 特徴

- クロンダイク（山札 1 枚めくり、再循環無制限）
- 7 列の場札・4 つの組札・山札・捨て札
- マウス / タッチでのドラッグ＆ドロップ（表向きの連なりはまとめて移動）。不正な移動はスナップバック
- ダブルクリック（タップ 2 回）で組札へ自動移動
- Undo（全手戻し可能）、新規ゲーム、同じ配牌でのやり直し
- 手数・経過時間・スコア（Windows ソリティア準拠の標準採点）
- 自動完了（山札・捨て札が空で場札が全て表向きのとき、1 手ずつアニメーションで組札へ）
- クリア演出（カードが跳ねる）
- seed による配牌の再現（URL クエリ `?seed=12345`）
- ウィンドウ / ブラウザ幅に応じたレイアウト（縦長画面にも対応）
- 日本語 UI（使用文字だけを事前生成したビットマップフォントのグリフアトラスを使用。フォント本体は同梱しない）

## 遊び方・操作

| 操作 | 内容 |
| --- | --- |
| カードをドラッグ | 移動（表向きの連なりはまとめて移動できます） |
| 山札をクリック / `Space` | 1 枚めくる。山札が尽きたら捨て札を山札へ戻す（−100 点） |
| カードをダブルクリック | 可能なら組札へ自動移動 |
| `N` / 「新規」ボタン | 新しい配牌でゲーム開始 |
| `R` / 「やり直し」ボタン | 同じ配牌（同じ seed）でやり直し |
| `Ctrl`+`Z` / 「戻す」ボタン | 1 手戻す（何手でも戻せます） |
| `A` / 「自動完了」ボタン | 条件を満たすとき残りを自動で組札へ |

### スコア

| 操作 | 得点 |
| --- | --- |
| 捨て札 → 場札 | +5 |
| 場札 → 組札 / 捨て札 → 組札 | +10 |
| 場札の裏向きカードをめくる | +5 |
| 組札 → 場札 | −15 |
| 山札の再循環 | −100 |

スコアの下限は 0 です。

### seed

画面左上に表示される seed をブラウザの URL（`?seed=12345`）やデスクトップ版の `-seed 12345` に指定すると、同じ配牌を再現できます。

## ビルドと実行

Go 1.26 以降が必要です。

```sh
# デスクトップ版
go run ./cmd/solitaire            # -seed 12345 で配牌を指定可能

# テスト
go vet ./... && go test ./...

# WebAssembly 版をビルドして http://127.0.0.1:8080/ で配信
make serve
# または Windows: .\build.ps1 serve（go が PATH に無い場合は $env:GO に go.exe のパスを指定）
```

`make wasm` / `.\build.ps1 wasm` は `web/solitaire.wasm` と `web/wasm_exec.js` を生成します（いずれも git 管理外）。

### ブラウザでの自動検証（任意）

```sh
make serve                      # 別ターミナルで配信を開始
npm i --no-save playwright-core
node tools/e2e/e2e.js           # BROWSER_PATH=... で Chromium 系ブラウザを指定可能
```

ランディングページの属性、canvas の描画、山札クリック / Space / Undo / ドラッグ＆ドロップ / ダブルクリック / 新規ゲーム時の URL 更新を headless ブラウザで確認し、コンソールエラーが 0 件であることを検証します。

## デプロイ（GitHub Pages）

`main` ブランチへ push すると [`.github/workflows/deploy.yml`](.github/workflows/deploy.yml) が
`gofmt` / `go vet` / `go test` → WASM ビルド → `web/` を Pages へデプロイします。

初回のみ、リポジトリの **Settings → Pages → Build and deployment → Source** を **GitHub Actions** に設定してください。

## 構成

```
cmd/solitaire/        エントリポイント（デスクトップ / WASM 共通）
internal/klondike/    ルール純ロジック（Ebitengine 非依存）。配牌・移動判定・Undo・スコア・勝利判定
internal/ui/          描画・入力・アニメーション・レイアウト
internal/assets/      埋め込み画像のロード
internal/cardfiles/   スート・ランク → ファイル名のマッピング（画像差し替え時はここを変更）
internal/uifont/      生成済みフォントアトラスと描画
internal/platform/    ブラウザ / デスクトップ差分（URL の seed、デバッグ API）
assets/cards/         縮小済みカード画像（embed）
tools/resize/         元画像からカード画像を縮小生成
tools/fontgen/        UI 文言からフォントアトラスを生成
tools/serve/          開発用静的サーバ
tools/e2e/            headless ブラウザによる検証スクリプト
web/                  index.html / game.html / main.js / style.css（Pages の公開ディレクトリ）
```

### カード画像の再生成

元画像（712×1008px PNG）を `golang.org/x/image/draw` の CatmullRom で縮小して `assets/cards/` に配置します。

```sh
go run ./tools/resize -src /path/to/torannpu -dst assets/cards -width 180
```

### フォントアトラスの再生成

UI の文言（`internal/ui/*.go` の文字列リテラル）に文字を追加したら、アトラスを再生成します。
不足があれば `internal/uifont` のテストが検出します。

```sh
go run ./tools/fontgen -src internal/ui -out internal/uifont
```

## 画像出典

カード画像の出典・ライセンス: **（記入してください）**

## フォント・ライセンス

UI 文字の描画には [bitmapfont/v4](https://github.com/hajimehoshi/bitmapfont)（Apache License 2.0）の 12px ビットマップフォント `FaceEA` から
`tools/fontgen` で事前生成したグリフアトラス（`internal/uifont/atlas.png`、使用文字のみ）を使っています。
元フォント（Ark Pixel Font / M+ Bitmap Font / misc-fixed / Baekmuk Gulim ほか）のライセンスは
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) を参照してください。

## ライセンス

本リポジトリのソースコードのライセンスは未定です（記入してください）。カード画像は上記「画像出典」に従います。
