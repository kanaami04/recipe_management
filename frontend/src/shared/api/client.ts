import { client } from './generated/client.gen'

// 生成 API クライアントの実行時設定 (frontend ADR-0007)。
// baseURL はハードコードせず env から取り、dev は Vite proxy 経由で /api を同一オリジン化する
// (ADR-0009 / ADR-0004)。アプリ起動時に一度呼ぶ。
// 認証(access 付与・401→refresh)の interceptor は ADR-0004 の段階でここに追加する。
export function configureApiClient() {
  client.setConfig({
    baseURL: import.meta.env.VITE_API_BASE_URL ?? '',
  })
}
