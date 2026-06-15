import { render, type RenderOptions } from '@testing-library/react'
import type { ReactElement, ReactNode } from 'react'
import { BrowserRouter } from 'react-router-dom'

// アプリ全体の Provider をまとめて差し込むテストヘルパー (ADR-0008)。
// 現状は Router のみ。TanStack Query の QueryClientProvider (ADR-0003) と
// 認証 Context (ADR-0004) は、それらが導入される段階でここへ追加する。
function AllProviders({ children }: { children: ReactNode }) {
  return <BrowserRouter>{children}</BrowserRouter>
}

export function renderWithProviders(ui: ReactElement, options?: Omit<RenderOptions, 'wrapper'>) {
  return render(ui, { wrapper: AllProviders, ...options })
}

// テストから RTL の API をこのモジュール経由で使えるよう再エクスポートする。
export * from '@testing-library/react'
export { default as userEvent } from '@testing-library/user-event'
