import { LoginPage } from '@/pages/LoginPage'

// 薄いルート: 描画は feature 側に委譲する (ADR-0002)。
export default function LoginRoute() {
  return <LoginPage />
}
