import { setupServer } from 'msw/node'

import { handlers } from './handlers'

// テスト用の MSW サーバ。ネットワーク層で intercept するため、
// axios interceptor もそのまま通る (ADR-0008)。
export const server = setupServer(...handlers)
