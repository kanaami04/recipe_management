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
      // dev は /api を Go バックエンドへ転送し同一オリジン化する (ADR-0004 / ADR-0009)。
      proxy: {
        '/api': {
          target: env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8000',
          changeOrigin: true,
        },
      },
    },
    // テスト基盤 (ADR-0008): Vitest + Testing Library + MSW。
    test: {
      environment: 'jsdom',
      setupFiles: ['./src/test/setup.ts'],
      css: false,
    },
  }
})
