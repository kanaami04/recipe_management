import { QueryClient } from '@tanstack/react-query'

// アプリ共通の QueryClient。
// サーバ状態のキャッシュ・再取得・無効化をここに集約する。
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
      staleTime: 30_000,
    },
  },
})
