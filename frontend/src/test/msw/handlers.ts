import type { RequestHandler } from 'msw'

// MSW のデフォルトハンドラ。
// 各 feature の API モックは、テスト側で server.use() を使うか、
// feature 配下にハンドラを定義してここへ集約する (ADR-0008)。
export const handlers: RequestHandler[] = []
