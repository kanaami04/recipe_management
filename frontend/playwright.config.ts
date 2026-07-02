import { defineConfig, devices } from '@playwright/test'

// E2E テスト (frontend ADR-0008 #4)。主要フロー(ログイン→一覧→作成)を検証する。
// API はブラウザ側のルートモックで差し替えるため、バックエンド/DB は不要。
export default defineConfig({
  testDir: './e2e',
  // E2E は dev サーバ共有 + メモリトークンのため直列実行で安定させる。
  fullyParallel: false,
  workers: 1,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  reporter: 'list',
  use: {
    baseURL: 'http://localhost:4173',
    trace: 'on-first-retry',
    // 本番ビルドには Service Worker が含まれる (ADR-0010)。SW がリクエストを
    // 横取りすると page.route のモックが不安定になるため、既定ではブロックする。
    // SW 自体の検証(e2e/pwa.spec.ts)はテスト側で 'allow' に上書きする。
    serviceWorkers: 'block',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  // 本番ビルド + preview を使う。dev の依存最適化リロードに起因する初回フレークを避ける。
  // API はブラウザ側のルートモックで差し替えるため proxy は不要。
  webServer: {
    command: 'npm run build && npm run preview -- --port 4173',
    url: 'http://localhost:4173',
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
})
