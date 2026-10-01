// Headless ブラウザ（Edge / Chrome / Chromium）で WASM 版ソリティアを検証する。
//
// 使い方（別ターミナルで `go run ./tools/serve` を起動しておく）:
//   npm i --no-save playwright-core
//   node tools/e2e/e2e.js
// 環境変数:
//   BROWSER_PATH  Chromium 系ブラウザの実行ファイル（既定: Windows の Edge）
//   BASE          配信 URL（既定: http://127.0.0.1:8080）
//   SHOT          指定するとゲーム画面のスクリーンショット PNG を保存する
const { chromium } = require("playwright-core");

const BROWSER_PATH = process.env.BROWSER_PATH || "C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe";
const BASE = process.env.BASE || "http://127.0.0.1:8080";

(async () => {
  const browser = await chromium.launch({ executablePath: BROWSER_PATH, headless: true, args: ["--use-gl=swiftshader", "--enable-unsafe-swiftshader"] });
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const page = await ctx.newPage();
  const errors = [];
  page.on("console", (m) => { if (m.type() === "error") errors.push("console: " + m.text() + " @ " + (m.location() && m.location().url)); });
  page.on("response", (r) => { if (r.status() >= 400) errors.push("http " + r.status() + ": " + r.url()); });
  page.on("pageerror", (e) => errors.push("pageerror: " + e.message));
  page.on("requestfailed", (r) => {
    const reason = r.failure() ? r.failure().errorText : "";
    if (reason === "net::ERR_ABORTED") return; // ページ遷移で中断されたリクエスト
    errors.push("requestfailed: " + r.url() + " " + reason);
  });

  // 1) ランディングページ（index.html）: iframe と a11y 要素
  await page.goto(BASE + "/?seed=12345", { waitUntil: "load" });
  const lang = await page.getAttribute("html", "lang");
  const title = await page.title();
  const iframeSrc = await page.getAttribute("#game", "src");
  const iframeTitle = await page.getAttribute("#game", "title");
  console.log(JSON.stringify({ lang, title, iframeSrc, iframeTitle }));

  const frame = page.frames().find((f) => f.url().includes("game.html"));
  if (!frame) throw new Error("game frame not found");
  await frame.waitForFunction(() => typeof window.solitaireState === "function", null, { timeout: 60000 });
  await frame.waitForFunction(() => document.querySelector("canvas") && !document.getElementById("loading"), null, { timeout: 30000 });
  const ariaLabel = await frame.evaluate(() => document.querySelector("canvas").getAttribute("aria-label"));
  const state0 = await frame.evaluate(() => window.solitaireState());
  console.log("aria-label:", ariaLabel);
  console.log("state0:", JSON.stringify(state0));

  // 2) ゲーム画面単体（game.html）: クリック操作で手数変化
  await page.goto(BASE + "/game.html?seed=12345", { waitUntil: "load" });
  await page.waitForFunction(() => typeof window.solitaireState === "function", null, { timeout: 60000 });
  await page.waitForTimeout(500);
  const s1 = await page.evaluate(() => window.solitaireState());
  console.log("s1:", JSON.stringify(s1));
  if (s1.seed !== 12345) throw new Error("seed not applied: " + s1.seed);
  if (s1.moves !== 0) throw new Error("moves should be 0");

  // canvas が描画されているか（中央付近に緑があるか）
  const shot = await page.screenshot({ type: "png" });
  const pixel = await page.evaluate(() => {
    const c = document.querySelector("canvas");
    return { w: c.width, h: c.height, cssW: c.clientWidth, cssH: c.clientHeight };
  });
  console.log("canvas:", JSON.stringify(pixel), "screenshot bytes:", shot.length);

  // 山札クリック
  const r = s1.stockRect;
  const scale = pixel.cssW / s1.logicalW;
  const cx = (r.x + r.w / 2) * scale, cy = (r.y + r.h / 2) * scale;
  await page.mouse.click(cx, cy);
  await page.waitForTimeout(300);
  const s2 = await page.evaluate(() => window.solitaireState());
  console.log("after stock click:", JSON.stringify({ moves: s2.moves, stock: s2.stock, waste: s2.waste }));
  if (s2.moves !== 1 || s2.waste !== 1) throw new Error("stock click did not draw");

  // Space で 2 枚目
  await page.keyboard.press("Space");
  await page.waitForTimeout(300);
  const s3 = await page.evaluate(() => window.solitaireState());
  console.log("after Space:", JSON.stringify({ moves: s3.moves, stock: s3.stock, waste: s3.waste }));
  if (s3.moves !== 2) throw new Error("Space did not draw");

  // Ctrl+Z で Undo
  await page.keyboard.press("Control+KeyZ");
  await page.waitForTimeout(300);
  const s4 = await page.evaluate(() => window.solitaireState());
  console.log("after Undo:", JSON.stringify({ moves: s4.moves, stock: s4.stock, waste: s4.waste }));
  if (s4.moves !== 1) throw new Error("Undo failed");

  // 「戻す」ボタンで Undo
  const ub = s4.undoButton;
  await page.mouse.click((ub.x + ub.w / 2) * scale, (ub.y + ub.h / 2) * scale);
  await page.waitForTimeout(300);
  const s4b = await page.evaluate(() => window.solitaireState());
  console.log("after Undo button:", JSON.stringify({ moves: s4b.moves, canUndo: s4b.canUndo }));
  if (s4b.moves !== 0 || s4b.canUndo) throw new Error("Undo button failed");

  // 不正なドラッグ（3♠ を 7♦ の上へ）→ スナップバック（手数不変）
  const c = (r) => [(r.x + r.w / 2) * scale, (r.y + r.h / 2) * scale];
  let [ax, ay] = c(s4b.tableauTops[0]);
  let [bx, by] = c(s4b.tableauTops[2]);
  await page.mouse.move(ax, ay); await page.mouse.down();
  for (let i = 1; i <= 10; i++) { await page.mouse.move(ax + (bx - ax) * i / 10, ay + (by - ay) * i / 10); await page.waitForTimeout(16); }
  await page.mouse.up();
  await page.waitForTimeout(600);
  const s6 = await page.evaluate(() => window.solitaireState());
  console.log("after invalid drag:", JSON.stringify({ moves: s6.moves, score: s6.score }));
  if (s6.moves !== 0) throw new Error("invalid drag should snap back");

  // 正しいドラッグ（A♣ を組札へ）→ +10
  [ax, ay] = c(s6.tableauTops[1]);
  [bx, by] = c(s6.foundations[0]);
  await page.mouse.move(ax, ay); await page.mouse.down();
  for (let i = 1; i <= 10; i++) { await page.mouse.move(ax + (bx - ax) * i / 10, ay + (by - ay) * i / 10); await page.waitForTimeout(16); }
  await page.mouse.up();
  await page.waitForTimeout(600);
  const s7 = await page.evaluate(() => window.solitaireState());
  console.log("after drag A to foundation:", JSON.stringify({ moves: s7.moves, score: s7.score, foundation: s7.foundation }));
  if (s7.foundation !== 1 || s7.score !== 15) throw new Error("drag to foundation failed (expected +10 move, +5 flip)");

  // Undo して、ダブルクリックで組札へ
  await page.keyboard.press("Control+KeyZ");
  await page.waitForTimeout(300);
  [ax, ay] = c(s6.tableauTops[1]);
  // Playwright の dblclick は 2 クリックが同一 tick に収まり 1 回の押下に潰れるため、実際の操作に近い間隔で 2 回クリックする。
  await page.mouse.click(ax, ay);
  await page.waitForTimeout(120);
  await page.mouse.click(ax, ay);
  await page.waitForTimeout(600);
  const s8 = await page.evaluate(() => window.solitaireState());
  console.log("after dblclick:", JSON.stringify({ moves: s8.moves, score: s8.score, foundation: s8.foundation }));
  if (s8.foundation !== 1) throw new Error("double click to foundation failed");

  // URL に seed が反映されているか / N で新規
  await page.keyboard.press("KeyN");
  await page.waitForTimeout(300);
  const s5 = await page.evaluate(() => window.solitaireState());
  const url = page.url();
  console.log("after N:", JSON.stringify({ seed: s5.seed, moves: s5.moves, url }));
  if (!url.includes("seed=" + s5.seed)) throw new Error("URL seed not updated: " + url);

  if (process.env.SHOT) {
    require("fs").writeFileSync(process.env.SHOT, shot);
    console.log("screenshot saved:", process.env.SHOT);
  }

  console.log("console errors:", errors.length, errors);
  await browser.close();
  if (errors.length) process.exit(1);
  console.log("E2E OK");
})().catch((e) => { console.error("E2E FAILED:", e); process.exit(1); });
