export type LoginUser = {
  username: string
  email: string
  password: string
}

export type Token = string | null

export type UserContextType = {
  user: LoginUser | null
  token: Token
  setToken: (token: Token) => void
}
