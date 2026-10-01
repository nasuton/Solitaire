<#
.SYNOPSIS
    Windows 向けの Makefile 相当スクリプト。

.DESCRIPTION
    Go を WebAssembly にビルドして web/ に配置し、開発用サーバで配信します。
    go が PATH に無い場合は環境変数 GO にパスを指定してください
    （例: $env:GO = "$HOME\sdk\go1.26.3\bin\go.exe"）。

.PARAMETER Task
    wasm    - web/solitaire.wasm をビルドし wasm_exec.js をコピーする（既定）
    serve   - wasm + http://127.0.0.1:8080/ で配信
    desktop - デスクトップ版 solitaire.exe をビルド
    test    - gofmt / go vet / go test / wasm 向け vet
    assets  - 元画像からカード画像を再生成（-Src で元画像フォルダを指定）
    fontgen - UI 文言からフォントアトラスを再生成
    clean   - 生成物を削除

.EXAMPLE
    .\build.ps1 serve
    .\build.ps1 assets -Src D:\work\Go\torannpu
#>
[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [ValidateSet('wasm', 'serve', 'desktop', 'test', 'assets', 'fontgen', 'clean')]
    [string]$Task = 'wasm',
    [string]$Src = '..\torannpu',
    [int]$Width = 180
)

$ErrorActionPreference = 'Stop'
$root = $PSScriptRoot
$webDir = Join-Path $root 'web'
$wasmOut = Join-Path $webDir 'solitaire.wasm'
$wasmExecOut = Join-Path $webDir 'wasm_exec.js'

function Resolve-Go {
    if ($env:GO -and (Test-Path $env:GO)) { return $env:GO }
    $cmd = Get-Command go -CommandType Application -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($cmd) { return $cmd.Source }
    throw 'go が見つかりません。PATH に追加するか $env:GO に go.exe のパスを設定してください。'
}

$go = Resolve-Go
$gofmt = Join-Path (Split-Path $go) 'gofmt.exe'

function Invoke-Checked {
    param([string]$Command, [string[]]$Arguments)
    Write-Host ">> $Command $($Arguments -join ' ')" -ForegroundColor Cyan
    & $Command @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "'$Command $($Arguments -join ' ')' failed with exit code $LASTEXITCODE"
    }
}

function Invoke-WithWasmEnv {
    param([scriptblock]$Body)
    $prevGoos = $env:GOOS
    $prevGoarch = $env:GOARCH
    try {
        $env:GOOS = 'js'
        $env:GOARCH = 'wasm'
        & $Body
    } finally {
        $env:GOOS = $prevGoos
        $env:GOARCH = $prevGoarch
    }
}

function Build-Wasm {
    $goroot = (& $go env GOROOT).Trim()
    # Go 1.24 以降は lib/wasm、それ以前は misc/wasm。
    $wasmExecSrc = @(
        (Join-Path $goroot 'lib\wasm\wasm_exec.js'),
        (Join-Path $goroot 'misc\wasm\wasm_exec.js')
    ) | Where-Object { Test-Path $_ } | Select-Object -First 1
    if (-not $wasmExecSrc) {
        throw "wasm_exec.js が $goroot 配下に見つかりません。"
    }

    Push-Location $root
    try {
        Invoke-WithWasmEnv {
            Invoke-Checked $go @('build', '-trimpath', '-ldflags=-s -w', '-o', $wasmOut, './cmd/solitaire')
        }
    } finally {
        Pop-Location
    }
    Copy-Item $wasmExecSrc $wasmExecOut -Force
    Get-Item $wasmOut, $wasmExecOut | Format-Table Name, Length, LastWriteTime -AutoSize
}

Push-Location $root
try {
    switch ($Task) {
        'wasm' { Build-Wasm }
        'serve' {
            Build-Wasm
            Invoke-Checked $go @('run', './tools/serve', '-dir', 'web', '-addr', '127.0.0.1:8080')
        }
        'desktop' {
            Invoke-Checked $go @('build', '-o', 'solitaire.exe', './cmd/solitaire')
        }
        'test' {
            $unformatted = & $gofmt -l .
            if ($unformatted) { throw "gofmt: 未整形のファイル:`n$($unformatted -join "`n")" }
            Invoke-Checked $go @('vet', './...')
            Invoke-WithWasmEnv { Invoke-Checked $go @('vet', './cmd/...', './internal/...') }
            Invoke-Checked $go @('test', './...')
        }
        'assets' {
            Invoke-Checked $go @('run', './tools/resize', '-src', $Src, '-dst', 'assets/cards', '-width', "$Width")
        }
        'fontgen' {
            Invoke-Checked $go @('run', './tools/fontgen', '-src', 'internal/ui', '-out', 'internal/uifont')
        }
        'clean' {
            foreach ($p in @($wasmOut, $wasmExecOut, (Join-Path $root 'solitaire.exe'))) {
                if (Test-Path $p) { Remove-Item $p -Force }
            }
        }
    }
} finally {
    Pop-Location
}
