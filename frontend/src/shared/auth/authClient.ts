import axios from 'axios'

import { CSRF_HEADER, CSRF_HEADER_VALUE } from './csrf'
import { clearAccessToken, setAccessToken } from './tokenStore'

// 認証専用の axios インスタンス (frontend ADR-0004)。
// auth interceptor は付けない(refresh が 401 で再帰しないようにするため)。
// refresh Cookie を送受信するため withCredentials。CSRF カスタムヘッダは固定付与する。
const authAxios = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '',
  withCredentials: true,
  headers: { [CSRF_HEADER]: CSRF_HEADER_VALUE },
})

// ログイン。access をメモリへ保持する。refresh はサーバが httpOnly Cookie で発行する。
export async function login(username: string, password: string): Promise<void> {
  const res = await authAxios.post('/api/token/', { username, password })
  setAccessToken(res.data.access)
}

// Cookie の refresh から新しい access を取得し保持する。
export async function refreshAccessToken(): Promise<string> {
  const res = await authAxios.post('/api/token/refresh/')
  const access: string = res.data.access
  setAccessToken(access)
  return access
}

// ログアウト。サーバ側で Cookie を失効させ、メモリの access を消す。
export async function logout(): Promise<void> {
  try {
    await authAxios.post('/api/auth/logout/')
  } finally {
    clearAccessToken()
  }
}
