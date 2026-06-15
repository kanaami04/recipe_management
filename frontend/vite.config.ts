import { reactRouter } from '@react-router/dev/vite'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import { loadEnv } from 'vite'
import { defineConfig } from 'vitest/config'

// Vitest 実行時は framework mode プラグイン(reactRouter)を使わず、
// @vitejs/plugin-react で JSX を変換する(reactRouter はテストランナーと相性が悪いため)。
const isTest = !!process.env.VITEST

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [tailwindcss(), ...(isTest ? [react()] : [reactRouter()])],
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
      // dev は /api を Go バックエンドへ転送し同一オリジン化する (ADR-0004 / ADR-0009)。
      proxy: {
        '/api': {
          target: env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8000',
          changeOrigin: true,
        },
      },
    },
    // テスト基盤 (ADR-0008): Vitest + Testing Library + MSW。
    // E2E(e2e/ 配下の Playwright)は Vitest の対象外にする。
    test: {
      environment: 'jsdom',
      setupFiles: ['./src/test/setup.ts'],
      css: false,
      include: ['src/**/*.{test,spec}.{ts,tsx}'],
    },
  }
})
