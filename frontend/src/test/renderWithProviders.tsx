import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, type RenderOptions } from '@testing-library/react'
import { type ReactElement, type ReactNode, useState } from 'react'
import { BrowserRouter } from 'react-router-dom'

// アプリ全体の Provider をまとめて差し込むテストヘルパー (ADR-0008)。
// QueryClient はテストごとに新規生成して状態を隔離する。
// 認証 Context (ADR-0004) は必要になった段階でここへ追加する。
function AllProviders({ children }: { children: ReactNode }) {
  const [client] = useState(
    () => new QueryClient({ defaultOptions: { queries: { retry: false } } }),
  )
  return (
    <QueryClientProvider client={client}>
      <BrowserRouter>{children}</BrowserRouter>
    </QueryClientProvider>
  )
}

export function renderWithProviders(ui: ReactElement, options?: Omit<RenderOptions, 'wrapper'>) {
  return render(ui, { wrapper: AllProviders, ...options })
}

// テストから RTL の API をこのモジュール経由で使えるよう再エクスポートする。
export * from '@testing-library/react'
export { default as userEvent } from '@testing-library/user-event'
