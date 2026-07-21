import { PasswordResetRequestPage } from '@/features/auth/components/PasswordResetRequestPage'

// 薄いルート: 描画は feature 側に委譲する。
export default function ResetPasswordRoute() {
  return <PasswordResetRequestPage />
}
