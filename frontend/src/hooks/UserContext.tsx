import { createContext, useContext, useEffect, useState } from 'react'

import { useFetchLoginUser } from '@/hooks/useFetchData'
import { getAccessToken, setAccessToken, subscribeAccessToken } from '@/shared/auth/tokenStore'
import type { Token, UserContextType } from '@/type/LoginUser'

const UserContext = createContext<UserContextType | null>(null)

// token はメモリストア(ADR-0004)を単一の源とし、Context は表示用にミラーする。
export const UserProvider = ({ children }: { children: React.ReactNode }) => {
  const [token, setTokenState] = useState<Token>(getAccessToken())

  useEffect(() => subscribeAccessToken(setTokenState), [])

  const { data: user } = useFetchLoginUser(token)

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
