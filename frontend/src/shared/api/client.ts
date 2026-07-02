import { attachAuthInterceptors } from '@/shared/auth/interceptors'

import { client } from './generated/client.gen'

// 生成 API クライアントの実行時設定。
// baseURL は env から(dev は Vite proxy 経由で /api 同一オリジン化)。
// withCredentials で refresh Cookie を送受信し、auth interceptor で
// access 自動付与・401→refresh・失敗ログアウトを行う。アプリ起動時に一度呼ぶ。
export function configureApiClient() {
  client.setConfig({
    baseURL: import.meta.env.VITE_API_BASE_URL ?? '',
    withCredentials: true,
  })
  attachAuthInterceptors(client.instance)
}
