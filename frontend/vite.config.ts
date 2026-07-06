import { reactRouter } from '@react-router/dev/vite'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import { loadEnv } from 'vite'
import { VitePWA } from 'vite-plugin-pwa'
import { defineConfig } from 'vitest/config'

// Vitest 実行時は framework mode プラグイン(reactRouter)を使わず、
// @vitejs/plugin-react で JSX を変換する(reactRouter はテストランナーと相性が悪いため)。
const isTest = !!process.env.VITEST

// PWA 設定。React Router v7 framework mode は index.html をプラグイン実行後に
// 生成するため、公式インテグレーションが出るまで 3 点のワークアラウンドを併用する:
// ① index.html を additionalManifestEntries で明示 precache
// ② SW 登録は自動注入でなく手動(src/pwa.ts)
// ③ manifest/アイコンの <link> は root.tsx に手書き
const pwa = () =>
  VitePWA({
    // 更新プロンプト UI を作らず、新 SW を即時有効化する最小構成。
    registerType: 'autoUpdate',
    injectRegister: false,
    // framework mode の client 出力先。既定(dist)のままだと sw.js が迷子になる。
    outDir: 'build/client',
    manifest: {
      name: 'レシピ管理',
      short_name: 'レシピ',
      description: '自分のレシピを登録・検索できる管理アプリ',
      lang: 'ja',
      display: 'standalone',
      orientation: 'portrait',
      start_url: '/',
      scope: '/',
      theme_color: '#ffffff',
      background_color: '#ffffff',
      icons: [
        { src: '/pwa-192x192.png', sizes: '192x192', type: 'image/png' },
        { src: '/pwa-512x512.png', sizes: '512x512', type: 'image/png' },
        {
          src: '/maskable-icon-512x512.png',
          sizes: '512x512',
          type: 'image/png',
          purpose: 'maskable',
        },
      ],
    },
    workbox: {
      globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
      // ワークアラウンド①: ビルド時点で index.html が存在しないため明示 precache する。
      // revision はビルド毎に変える(アセット参照が変わるので実質毎ビルド更新される)。
      additionalManifestEntries: [{ url: '/index.html', revision: Date.now().toString(36) }],
      navigateFallback: '/index.html',
      // /api は SW に一切触らせない(認証レスポンスをキャッシュする事故の防止)。
      navigateFallbackDenylist: [/^\/api\//],
      cleanupOutdatedCaches: true,
    },
    // dev では SW を動かさない(既存の dev/E2E フローに影響させない)。
    devOptions: { enabled: false },
  })

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [tailwindcss(), ...(isTest ? [react()] : [reactRouter(), pwa()])],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    server: {
      // 他プロジェクト(Vite 既定の 5173)と衝突しないよう専用ポートに固定する。
      // strictPort で勝手にずらさせない(スマホ実機テストの URL を一定に保つ)。
      port: 5273,
      strictPort: true,
      // dev は /api を Go バックエンドへ転送し同一オリジン化する。
      proxy: {
        '/api': {
          target: env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8000',
          changeOrigin: true,
        },
      },
    },
    // テスト基盤: Vitest + Testing Library + MSW。
    // E2E(e2e/ 配下の Playwright)は Vitest の対象外にする。
    test: {
      environment: 'jsdom',
      setupFiles: ['./src/test/setup.ts'],
      css: false,
      include: ['src/**/*.{test,spec}.{ts,tsx}'],
    },
  }
})
