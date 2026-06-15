import type { AxiosError, AxiosInstance, InternalAxiosRequestConfig } from 'axios'

import { refreshAccessToken } from './authClient'
import { CSRF_HEADER, CSRF_HEADER_VALUE } from './csrf'
import { clearAccessToken, getAccessToken } from './tokenStore'

let refreshPromise: Promise<string> | null = null

// single-flight: 同時多発の 401 で refresh が複数回走らないよう 1 回に集約する (ADR-0004)。
function refreshOnce(): Promise<string> {
  if (!refreshPromise) {
    refreshPromise = refreshAccessToken().finally(() => {
      refreshPromise = null
    })
  }
  return refreshPromise
}

type RetriableConfig = InternalAxiosRequestConfig & { _retried?: boolean }

// axios インスタンスに認証 interceptor を付ける (ADR-0004)。
// - リクエスト: メモリの access を Authorization に自動付与する。
// - レスポンス: 401 を受けたら refresh で更新 → 元リクエストを 1 度だけ再試行 →
//   refresh も失敗ならログアウト(ストアを消して "/" へ)。
export function attachAuthInterceptors(instance: AxiosInstance): void {
  instance.interceptors.request.use((config) => {
    const token = getAccessToken()
    if (token) {
      config.headers.set('Authorization', `Bearer ${token}`)
    }
    // CSRF 対策のカスタムヘッダを全リクエストに付与する (ADR-0004 #3)。
    config.headers.set(CSRF_HEADER, CSRF_HEADER_VALUE)
    return config
  })

  instance.interceptors.response.use(
    (response) => response,
    async (error: AxiosError) => {
      const original = error.config as RetriableConfig | undefined
      if (error.response?.status === 401 && original && !original._retried) {
        original._retried = true
        try {
          const access = await refreshOnce()
          original.headers.set('Authorization', `Bearer ${access}`)
          return instance(original)
        } catch {
          clearAccessToken()
          if (typeof window !== 'undefined') {
            window.location.assign('/')
          }
          return Promise.reject(error)
        }
      }
      return Promise.reject(error)
    },
  )
}
