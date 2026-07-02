# 0010. PWA 対応: インストール可能 + 静的アセット precache

- ステータス: Accepted
- 日付: 2026-07-02

## コンテキスト

スマホからの利用が主体のアプリなので、ホーム画面に追加して全画面(standalone)で
起動できるようにしたい。また 2 回目以降の起動を速くしたい。

一方で本アプリはログイン必須の private アプリであり([ADR-0001](0001-build-tooling-and-app-shape.md))、
データは常に API から取る設計([ADR-0003](0003-data-fetching-tanstack-query.md))。
オフラインでのデータ閲覧まで踏み込むと、認証(メモリ保持の access token、
[ADR-0004](0004-auth-and-token-management.md))との兼ね合いで設計コストが大きい。

また React Router v7 framework mode([ADR-0002](0002-routing-react-router-framework-mode.md))は
`index.html` を Vite プラグインの実行後に生成するため、vite-plugin-pwa の
公式サポートが存在しない(vite-pwa/vite-plugin-pwa#809, remix-run/react-router#14268)。

## 決定

### 1. スコープは「インストール可能 + 静的アセット precache」に絞る

- Web App Manifest + アイコン一式 + Service Worker(Workbox generateSW)を導入する。
- precache はビルド成果物(JS/CSS/HTML/画像)のみ。**API レスポンスは一切キャッシュしない**
  (`navigateFallbackDenylist: [/^\/api\//]`。認証レスポンスをキャッシュする事故を防ぐ)。
- オフラインでのデータ閲覧は対応しない(オフライン時はアプリシェルだけ表示され、API はエラーになる)。

### 2. vite-plugin-pwa を 3 点のワークアラウンド付きで使う(暫定)

framework mode との相性問題は以下で回避する。**公式対応が入ったら削除する暫定措置**。

| 問題 | ワークアラウンド |
|---|---|
| index.html がプラグイン実行後に生成され precache に入らない | `additionalManifestEntries` で `/index.html` を明示追加(revision はビルド毎に更新) |
| SW 登録スクリプトの自動注入が効かない | `src/pwa.ts` で手動登録(`root.tsx` から本番ビルドのみ動的 import) |
| manifest link の自動注入が効かない | `root.tsx` の head に `<link rel="manifest">` 等を手書き |

### 3. SW の更新は autoUpdate 方式

`registerType: 'autoUpdate'`(skipWaiting + clientsClaim)とし、更新プロンプト UI は作らない。
オフラインデータを持たないため、更新の競合リスクは実質なく、最小構成を優先する。

### 4. dev では SW を動かさず、既存 E2E は SW をブロックする

- `devOptions.enabled: false`。dev サーバと Vitest は従来どおり(プラグインも isTest 時は除外)。
- E2E は本番ビルドに対して実行するため([ADR-0008](0008-testing-vitest-rtl-msw.md))、SW が
  `page.route` のモックを横取りしないよう playwright.config.ts で `serviceWorkers: 'block'` を既定にする。
  SW 自体の検証は `e2e/pwa.spec.ts` がテスト単位で `'allow'` に上書きして行う。

### 5. アイコンは public/apple.png から生成する

`@vite-pwa/assets-generator`(preset: minimal-2023)で 64/192/512/maskable/apple-touch-icon/favicon を
生成しコミットする。元画像を差し替えたら再生成する。

## 結果

### 良い点

- ホーム画面に追加でき、standalone で起動する。静的アセットがキャッシュされ 2 回目以降の起動が速い。
- API を SW から完全に切り離したため、認証設計([ADR-0004](0004-auth-and-token-management.md))に影響がない。

### トレードオフ

- オフラインではデータが見られない(スコープ外とした)。
- framework mode 対応のワークアラウンド 3 点は暫定で、公式対応後に削除する保守負担が残る。
- SW 起因の不具合調査(古いアセットが残る等)という新しい故障モードが増える。
  `cleanupOutdatedCaches` と autoUpdate で緩和する。
