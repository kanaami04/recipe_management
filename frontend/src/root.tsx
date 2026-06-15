import './index.css'

import { isRouteErrorResponse, Links, Meta, Outlet, Scripts, ScrollRestoration } from 'react-router'
import { Toaster } from 'sonner'

import { UserProvider } from './hooks/UserContext'

// HTML ドキュメントの骨格。framework mode が描画する (ADR-0002)。
export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ja">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
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
    <UserProvider>
      <Outlet />
      {/* 通知系トースト。alert() を置き換える (ADR-0009 #3) */}
      <Toaster richColors position="top-center" />
    </UserProvider>
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
