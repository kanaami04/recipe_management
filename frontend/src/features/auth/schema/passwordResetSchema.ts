import { z } from 'zod'

// パスワードリセット申請フォーム。メールアドレスのみ。
export const passwordResetRequestSchema = z.object({
  email: z
    .string()
    .min(1, 'メールアドレスは必須です')
    .email('メールアドレスの形式が正しくありません')
    .max(50, '50文字以内で入力してください'),
})

export type PasswordResetRequestValues = z.infer<typeof passwordResetRequestSchema>

// パスワードリセット確定フォーム。新パスワードは 8 文字以上(サーバの min=8 に揃える)。
export const passwordResetConfirmSchema = z
  .object({
    password: z.string().min(8, 'パスワードは8文字以上にしてください'),
    confirmPassword: z.string(),
  })
  .refine((v) => v.password === v.confirmPassword, {
    path: ['confirmPassword'],
    message: 'パスワードが一致しません',
  })

export type PasswordResetConfirmValues = z.infer<typeof passwordResetConfirmSchema>
