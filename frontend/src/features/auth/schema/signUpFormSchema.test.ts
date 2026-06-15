import { describe, expect, it } from 'vitest'

import { signUpFormSchema } from './signUpFormSchema'

const valid = { username: 'taro', email: 'taro@example.com', password: 'password123' }

describe('signUpFormSchema', () => {
  it('妥当な入力の時、検証を通過すること。', () => {
    expect(signUpFormSchema.safeParse(valid).success).toBe(true)
  })

  it('メール形式が不正な時、検証に失敗すること。', () => {
    expect(signUpFormSchema.safeParse({ ...valid, email: 'not-an-email' }).success).toBe(false)
  })

  it('パスワードが8文字未満の時、検証に失敗すること。', () => {
    expect(signUpFormSchema.safeParse({ ...valid, password: 'short' }).success).toBe(false)
  })

  it('ユーザー名が空の時、検証に失敗すること。', () => {
    expect(signUpFormSchema.safeParse({ ...valid, username: '' }).success).toBe(false)
  })
})
