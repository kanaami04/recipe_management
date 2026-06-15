import type { AxiosInstance } from 'axios'
import axios from 'axios'

import type { Token } from '@/type/LoginUser'

// baseURL はハードコードせず環境変数から取る (ADR-0009 #1)。
// 既定は空文字 = 相対 /api。dev は Vite proxy 経由で同一オリジン化する (ADR-0004)。
export const api: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '',
  withCredentials: false, // JWTは手動で送る
})

export const fetcher = (url: string, token: Token) =>
  api
    .get(url, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .then((res) => res.data)
