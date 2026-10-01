# サードパーティ・ライセンス表記

## フォント（`internal/uifont/atlas.png`）

UI 文字列の描画に使っているグリフアトラスは、
[github.com/hajimehoshi/bitmapfont/v4](https://github.com/hajimehoshi/bitmapfont)（Apache License 2.0）の
`FaceEA`（12px ビットマップフォント）から `tools/fontgen` で事前生成したものです（使用文字のみを含みます）。
bitmapfont は以下のフォントを元にしています（bitmapfont の README より）。

- [Ark Pixel Font](https://ark-pixel-font.takwolf.com/) (OFL-1.1)
- [Baekmuk Gulim](https://kldp.net/baekmuk/) (Baekmuk License)
- [Cubic 11](https://github.com/ACh-K/Cubic-11) (OFL-1.1)
- [Gulmuri](https://quiple.dev/galmuri) (OFL-1.1)
- [misc-fixed](https://www.cl.cam.ac.uk/~mgk25/ucs-fonts.html) (Public Domain)
- [M+ Bitmap Font](https://github.com/coz-m/MPLUS_FONTS/tree/master/obsolete) (M+ Bitmap Fonts License)
- Arabic glyphs by [@MansourSorosoro](https://twitter.com/MansourSorosoro) (Eternal Dream Arabization) (OFL-1.1)

### Baekmuk License

```
Copyright (c) 1986-2002 Kim Jeong-Hwan
All rights reserved.

Permission to use, copy, modify and distribute this font is
hereby granted, provided that both the copyright notice and
this permission notice appear in all copies of the font,
derivative works or modified versions, and that the following
acknowledgement appear in supporting documentation:
    Baekmuk Batang, Baekmuk Dotum, Baekmuk Gulim, and
    Baekmuk Headline are registered trademarks owned by
    Kim Jeong-Hwan.
```

### M+ Bitmap Font License

```
-
M+ BITMAP FONTS            Copyright 2002-2005  COZ <coz@users.sourceforge.jp>
-

LICENSE

These fonts are free softwares.
Unlimited permission is granted to use, copy, and distribute it, with
or without modification, either commercially and noncommercially.
THESE FONTS ARE PROVIDED "AS IS" WITHOUT WARRANTY.
```

## ライブラリ

- [Ebitengine](https://github.com/hajimehoshi/ebiten) — Apache License 2.0
- [golang.org/x/image](https://pkg.go.dev/golang.org/x/image) — BSD-3-Clause（`tools/resize` の縮小処理で使用）

## カード画像

`assets/cards/` のカード画像の出典・ライセンスは [README.md](README.md) の「画像出典」を参照してください。
