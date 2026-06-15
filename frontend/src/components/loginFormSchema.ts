import { z } from 'zod'

// ログインフォームの zod スキーマ(手書き、ADR-0006)。
export const loginFormSchema = z.object({
  username: z.string().min(1, 'ユーザー名は必須です'),
  password: z.string().min(1, 'パスワードは必須です'),
})

export type LoginFormValues = z.infer<typeof loginFormSchema>
