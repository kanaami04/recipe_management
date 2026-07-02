import { setupServer } from 'msw/node'

import { handlers } from './handlers'

// テスト用の MSW サーバ。ネットワーク層で intercept するため、
// axios interceptor もそのまま通る。
export const server = setupServer(...handlers)
