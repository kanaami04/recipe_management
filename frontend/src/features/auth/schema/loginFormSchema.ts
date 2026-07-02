import { z } from 'zod'

// ログインフォームの zod スキーマ(手書き、ADR-0006)。
export const loginFormSchema = z.object({
  email: z
    .string()
    .min(1, 'メールアドレスは必須です')
    .email('メールアドレスの形式が正しくありません'),
  password: z.string().min(1, 'パスワードは必須です'),
})

export type LoginFormValues = z.infer<typeof loginFormSchema>
