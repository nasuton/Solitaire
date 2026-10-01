// Command serve は web/ ディレクトリを配信する開発用の静的サーバ。
// .wasm に正しい MIME タイプを付け、キャッシュを無効化する。
//
//	go run ./tools/serve -dir web -addr :8080
package main

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
)

func main() {
	dir := flag.String("dir", "web", "配信するディレクトリ")
	addr := flag.String("addr", "127.0.0.1:8080", "待ち受けアドレス")
	flag.Parse()

	if _, err := os.Stat(*dir); err != nil {
		log.Fatalf("serve: %v", err)
	}
	_ = mime.AddExtensionType(".wasm", "application/wasm")
	_ = mime.AddExtensionType(".js", "text/javascript")

	fs := http.FileServer(http.Dir(*dir))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		fs.ServeHTTP(w, r)
	})

	fmt.Printf("serving %s at http://%s/\n", *dir, *addr)
	log.Fatal(http.ListenAndServe(*addr, handler))
}
