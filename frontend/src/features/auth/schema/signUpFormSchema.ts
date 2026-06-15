import { z } from 'zod'

// サインアップフォームの zod スキーマ(手書き、ADR-0006)。
// バックエンドの RegisterRequest 検証(username max50 / email / password min8)に合わせる。
export const signUpFormSchema = z.object({
  username: z.string().min(1, 'ユーザー名は必須です').max(50, '50文字以内で入力してください'),
  email: z
    .string()
    .min(1, 'メールアドレスは必須です')
    .email('メールアドレスの形式が正しくありません')
    .max(50, '50文字以内で入力してください'),
  password: z.string().min(8, 'パスワードは8文字以上で入力してください'),
})

export type SignUpFormValues = z.infer<typeof signUpFormSchema>
