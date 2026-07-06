import { Link } from 'react-router-dom'

import { SignUpForm } from '@/features/auth/components/SignUpForm'
import { Logo } from '@/shared/components/Logo'

export function SignUpPage() {
  return (
    <div className="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
      <div className="flex w-full max-w-sm flex-col gap-6">
        <Link to="/" className="self-center">
          <Logo markClassName="size-7" wordmarkClassName="text-xl" />
        </Link>
        <SignUpForm />
      </div>
    </div>
  )
}
