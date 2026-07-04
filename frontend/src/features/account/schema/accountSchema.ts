import { z } from 'zod'

// プロフィール編集(ユーザー名)。メールは本人確認が要るため EmailForm で別扱い。
export const profileFormSchema = z.object({
  username: z.string().min(1, 'ユーザー名は必須です').max(50, '50文字以内で入力してください'),
})

export type ProfileFormValues = z.infer<typeof profileFormSchema>

// メールアドレス変更。本人確認のため現在のパスワードを要求する。
export const emailFormSchema = z.object({
  email: z
    .string()
    .min(1, 'メールアドレスは必須です')
    .email('メールアドレスの形式が正しくありません')
    .max(50, '50文字以内で入力してください'),
  password: z.string().min(1, 'パスワードを入力してください'),
})

export type EmailFormValues = z.infer<typeof emailFormSchema>

// パスワード変更。新しいパスワードは 8 文字以上(サーバの min=8 に揃える)。
export const passwordFormSchema = z
  .object({
    currentPassword: z.string().min(1, '現在のパスワードを入力してください'),
    newPassword: z.string().min(8, '新しいパスワードは8文字以上にしてください'),
    confirmPassword: z.string(),
  })
  .refine((v) => v.newPassword === v.confirmPassword, {
    path: ['confirmPassword'],
    message: 'パスワードが一致しません',
  })

export type PasswordFormValues = z.infer<typeof passwordFormSchema>
