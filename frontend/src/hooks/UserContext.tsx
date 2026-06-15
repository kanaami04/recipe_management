import { createContext, useContext } from 'react'
import { useState } from 'react'

import { useFetchLoginUser } from '@/hooks/useFetchData'
import type { UserContextType } from '@/type/LoginUser.tsx'
import type { Token } from '@/type/LoginUser.tsx'

const UserContext = createContext<UserContextType | null>(null)

export const UserProvider = ({ children }: { children: React.ReactNode }) => {
  const [token, setToken] = useState<Token>(localStorage.getItem('access'))
  const { data: user, error } = useFetchLoginUser(token)

  if (error) console.log(error)

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
