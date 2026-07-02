import { redirect } from 'react-router'

import { refreshAccessToken } from './authClient'
import { getAccessToken } from './tokenStore'

// 保護ルートの clientLoader ガード。
// access が無ければ Cookie の refresh で復帰を試み、失敗なら "/" へリダイレクトする。
// clientLoader は React の外側で動くため、トークンはストアから読む。
export async function requireAuth() {
  if (getAccessToken()) {
    return null
  }
  try {
    await refreshAccessToken()
    return null
  } catch {
    throw redirect('/')
  }
}
