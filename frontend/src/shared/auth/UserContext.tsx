import { useQuery } from '@tanstack/react-query'
import { createContext, useContext, useEffect, useState } from 'react'

import { getUserInfoOptions } from '@/shared/api/generated/@tanstack/react-query.gen'
import { getAccessToken, setAccessToken, subscribeAccessToken } from '@/shared/auth/tokenStore'
import type { Token, UserContextType } from '@/shared/auth/types'

const UserContext = createContext<UserContextType | null>(null)

// token はメモリストア(ADR-0004)を単一の源とし、Context は表示用にミラーする。
export const UserProvider = ({ children }: { children: React.ReactNode }) => {
  const [token, setTokenState] = useState<Token>(getAccessToken())

  useEffect(() => subscribeAccessToken(setTokenState), [])

  // ログインユーザー情報は生成 Query フックで取得する (ADR-0003/0007)。
  // token がある時だけ叩く。認証ヘッダは interceptor が付与する。
  const { data: user } = useQuery({ ...getUserInfoOptions(), enabled: !!token })

  const setToken = (next: Token) => setAccessToken(next)

  return (
    <UserContext.Provider value={{ user: user ?? null, token, setToken }}>
      {children}
    </UserContext.Provider>
  )
}

export const useUser = () => {
  const context = useContext(UserContext)
  if (!context) {
    throw new Error('useUser は UserProvider の中で使う必要があります')
  }
  return context
}
