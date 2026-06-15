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
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  // dev サーバを起動して E2E を回す。既存サーバがあれば再利用する。
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
})
