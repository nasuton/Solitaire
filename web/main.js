// solitaire.wasm を読み込んで起動する。
(function () {
  "use strict";

  function showError(message) {
    var loading = document.getElementById("loading");
    if (!loading) return;
    loading.textContent = "";
    var pre = document.createElement("pre");
    pre.id = "error";
    pre.textContent = "読み込みに失敗しました:\n" + message;
    loading.appendChild(pre);
  }

  if (typeof WebAssembly === "undefined" || typeof Go === "undefined") {
    showError("このブラウザは WebAssembly に対応していません。");
    return;
  }

  var go = new Go();
  var wasmURL = "solitaire.wasm";
  var promise = WebAssembly.instantiateStreaming
    ? WebAssembly.instantiateStreaming(fetch(wasmURL), go.importObject)
    : fetch(wasmURL)
        .then(function (r) { return r.arrayBuffer(); })
        .then(function (b) { return WebAssembly.instantiate(b, go.importObject); });

  promise
    .then(function (result) {
      return go.run(result.instance);
    })
    .catch(function (err) {
      console.error(err);
      showError(err && err.message ? err.message : String(err));
    });
})();
