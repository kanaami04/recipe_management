import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { Link } from 'react-router-dom'

import {
  passwordResetRequestSchema,
  type PasswordResetRequestValues,
} from '@/features/auth/schema/passwordResetSchema'
import { requestPasswordResetMutation } from '@/shared/api/generated/@tanstack/react-query.gen'
import { Logo } from '@/shared/components/Logo'
import { Button } from '@/shared/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'

// パスワードリセット申請ページ。メールアドレスを受け取り、リセットリンクを送る。
// メール列挙を防ぐため、成否に関わらず同じ「送信しました」表示にする。
export function PasswordResetRequestPage() {
  const [submitted, setSubmitted] = useState(false)

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<PasswordResetRequestValues>({
    resolver: zodResolver(passwordResetRequestSchema),
    defaultValues: { email: '' },
    mode: 'onBlur',
  })

  // サーバは存在有無に関わらず 204。成功・失敗どちらでも同じ完了表示にする。
  const requestReset = useMutation({
    ...requestPasswordResetMutation(),
    onSuccess: () => setSubmitted(true),
    onError: () => setSubmitted(true),
  })

  const onSubmit = handleSubmit((values) => {
    requestReset.mutate({ body: { email: values.email } })
  })

  return (
    <div className="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
      <div className="flex w-full max-w-sm flex-col gap-6">
        <Link to="/" className="self-center">
          <Logo markClassName="size-7" wordmarkClassName="text-xl" />
        </Link>
        <Card className="w-full max-w-sm">
          <CardHeader>
            <CardTitle>パスワードの再設定</CardTitle>
            <CardDescription>
              {submitted
                ? '入力されたメールアドレスが登録済みであれば、再設定用のリンクを送信しました。メールをご確認ください。'
                : '登録済みのメールアドレスを入力してください。再設定用のリンクを送ります。'}
            </CardDescription>
          </CardHeader>
          {submitted ? (
            <CardContent>
              <Button asChild variant="outline" className="w-full">
                <Link to="/">ログイン画面へ戻る</Link>
              </Button>
            </CardContent>
          ) : (
            <form onSubmit={onSubmit}>
              <CardContent>
                <div className="grid gap-2">
                  <Label htmlFor="email">メールアドレス</Label>
                  <Controller
                    control={control}
                    name="email"
                    render={({ field }) => (
                      <Input
                        id="email"
                        type="email"
                        value={field.value}
                        onChange={field.onChange}
                        onBlur={field.onBlur}
                      />
                    )}
                  />
                  {errors.email && (
                    <p className="text-destructive text-sm">{errors.email.message}</p>
                  )}
                </div>
              </CardContent>
              <CardFooter className="mt-6 flex-col gap-2">
                <Button type="submit" className="w-full" disabled={requestReset.isPending}>
                  再設定リンクを送る
                </Button>
                <Button asChild variant="link" className="w-full">
                  <Link to="/">ログイン画面へ戻る</Link>
                </Button>
              </CardFooter>
            </form>
          )}
        </Card>
      </div>
    </div>
  )
}
