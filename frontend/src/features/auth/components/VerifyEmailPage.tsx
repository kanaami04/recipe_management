import { useMutation } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'
import { Link, useSearchParams } from 'react-router-dom'

import { verifyEmailMutation } from '@/shared/api/generated/@tanstack/react-query.gen'
import { Logo } from '@/shared/components/Logo'
import { Button } from '@/shared/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'

// メール確認ページ。URL の ?token を検証 API に渡し、結果を表示する。
// 未ログインで到達するため、公開ルート(/verify-email)に置く。
export function VerifyEmailPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token') ?? ''

  const verify = useMutation(verifyEmailMutation())

  // トークンで一度だけ検証する。StrictMode の二重実行を ref で抑止する。
  const startedRef = useRef(false)
  useEffect(() => {
    if (startedRef.current || !token) return
    startedRef.current = true
    verify.mutate({ body: { token } })
  }, [token, verify])

  const status = !token
    ? 'error'
    : verify.isSuccess
      ? 'success'
      : verify.isError
        ? 'error'
        : 'pending'

  return (
    <div className="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
      <div className="flex w-full max-w-sm flex-col gap-6">
        <Link to="/" className="self-center">
          <Logo markClassName="size-7" wordmarkClassName="text-xl" />
        </Link>
        <Card className="w-full max-w-sm">
          <CardHeader>
            <CardTitle>
              {status === 'pending' && 'メールアドレスを確認しています'}
              {status === 'success' && '確認が完了しました'}
              {status === 'error' && '確認できませんでした'}
            </CardTitle>
            <CardDescription>
              {status === 'pending' && 'しばらくお待ちください。'}
              {status === 'success' && 'メールアドレスの確認が完了しました。ログインできます。'}
              {status === 'error' &&
                'リンクが無効か、有効期限(24時間)が切れています。ログイン画面から確認メールを再送してください。'}
            </CardDescription>
          </CardHeader>
          {status !== 'pending' && (
            <CardContent>
              <Button asChild className="w-full">
                <Link to="/">ログイン画面へ</Link>
              </Button>
            </CardContent>
          )}
        </Card>
      </div>
    </div>
  )
}
