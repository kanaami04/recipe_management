import type { UserInfoResponse } from '@/shared/api/generated/types.gen'

export type Token = string | null

export type UserContextType = {
  user: UserInfoResponse | null
  token: Token
  setToken: (token: Token) => void
}
