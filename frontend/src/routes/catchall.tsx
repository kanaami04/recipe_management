import { redirect } from 'react-router'

import { refreshAccessToken } from '@/shared/auth/authClient'
import { getAccessToken } from '@/shared/auth/tokenStore'

// 未定義パスのリダイレクト先を認証状態で振り分ける (ADR-0002/0004)。
// 認証済み(access あり、または Cookie の refresh で復帰可能)なら /top、
// 未認証ならログイン("/")へ。clientLoader は React の外側で動くためストアから読む。
export async function clientLoader() {
  if (getAccessToken()) {
    return redirect('/top')
  }
  try {
    await refreshAccessToken()
    return redirect('/top')
  } catch {
    return redirect('/')
  }
}

export default function CatchAll() {
  return null
}
