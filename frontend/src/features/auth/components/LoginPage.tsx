import { Link } from 'react-router-dom'

import { LoginForm } from '@/features/auth/components/LoginForm'
import { Logo } from '@/shared/components/Logo'

export function LoginPage() {
  return (
    <div className="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
      <div className="flex w-full max-w-sm flex-col gap-6">
        <Link to="/" className="self-center">
          <Logo markClassName="size-7" wordmarkClassName="text-xl" />
        </Link>
        <LoginForm />
      </div>
    </div>
  )
}
