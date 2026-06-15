import { SignUpPage } from '@/features/auth/components/SignUpPage'

// 薄いルート: 描画は feature 側に委譲する (ADR-0002)。
export default function SignUpRoute() {
  return <SignUpPage />
}
