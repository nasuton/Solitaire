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
- [golang.org/x/image](https://pkg.go.dev/golang.org/x/image) — BSD-3-Clause（`tools/cardgen` のベジェ描画・フォント描画で使用）

## カード画像（`assets/cards/*.png`）

カード画像は本リポジトリの `tools/cardgen` で生成した自作のもので、[MIT License](LICENSE) です。
ランク文字（A, 2–10, J, Q, K）の描画には [Go フォント](https://go.dev/blog/go-fonts) の `Go Bold`
（`golang.org/x/image/font/gofont/gobold`）をラスタライズして使用しています。フォント自体はリポジトリに同梱していません。

### Go Fonts License (BSD-3-Clause)

```
Copyright (c) 2016 Bigelow & Holmes Inc.. All rights reserved.

Distribution of this font is governed by the following license. If you do not
agree to this license, including the disclaimer, do not distribute or modify
this font.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

	* Redistributions of source code must retain the above copyright notice,
	  this list of conditions and the following disclaimer.

	* Redistributions in binary form must reproduce the above copyright notice,
	  this list of conditions and the following disclaimer in the documentation
	  and/or other materials provided with the distribution.

	* Neither the name of Google Inc. nor the names of its contributors may be
	  used to endorse or promote products derived from this software without
	  specific prior written permission.

DISCLAIMER: THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO,
THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```
