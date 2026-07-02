import './styles/index.css'

import { QueryClientProvider } from '@tanstack/react-query'
import { isRouteErrorResponse, Links, Meta, Outlet, Scripts, ScrollRestoration } from 'react-router'
import { Toaster } from 'sonner'

import { configureApiClient } from './shared/api/client'
import { UserProvider } from './shared/auth/UserContext'
import { queryClient } from './shared/lib/queryClient'

// 生成 API クライアントに baseURL・withCredentials・auth interceptor を適用する (ADR-0004/0007)。
configureApiClient()

// Service Worker の登録 (ADR-0010)。本番ビルドのブラウザ上のみ
// (dev/Vitest と、ビルド時の index.html プリレンダリングでは動かさない)。
if (import.meta.env.PROD && typeof document !== 'undefined') {
  import('./pwa')
}

// HTML ドキュメントの骨格。framework mode が描画する (ADR-0002)。
export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ja">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        {/* PWA (ADR-0010)。framework mode では vite-plugin-pwa の自動注入が
            効かないため manifest/アイコンを手書きする(ワークアラウンド③)。 */}
        <meta name="theme-color" content="#ffffff" />
        <link rel="icon" href="/favicon.ico" sizes="48x48" />
        <link rel="apple-touch-icon" href="/apple-touch-icon-180x180.png" />
        <link rel="manifest" href="/manifest.webmanifest" />
        <Meta />
        <Links />
      </head>
      <body>
        {children}
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  )
}

// アプリのルート。Provider を集約する(QueryClientProvider は ADR-0003 段階で追加)。
export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <UserProvider>
        <Outlet />
        {/* 通知系トースト。alert() を置き換える (ADR-0009 #3) */}
        <Toaster richColors position="top-center" />
      </UserProvider>
    </QueryClientProvider>
  )
}

// ルート単位のエラー捕捉 (ADR-0009 #6)。
export function ErrorBoundary({ error }: { error: unknown }) {
  let message = '予期しないエラーが発生しました。'
  if (isRouteErrorResponse(error)) {
    message = `${error.status} ${error.statusText}`
  }
  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-2 p-6">
      <h1 className="text-lg font-medium">エラー</h1>
      <p className="text-muted-foreground">{message}</p>
    </div>
  )
}
