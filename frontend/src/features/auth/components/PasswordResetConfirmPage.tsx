import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { Controller, useForm } from 'react-hook-form'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { toast } from 'sonner'

import {
  passwordResetConfirmSchema,
  type PasswordResetConfirmValues,
} from '@/features/auth/schema/passwordResetSchema'
import { confirmPasswordResetMutation } from '@/shared/api/generated/@tanstack/react-query.gen'
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

// パスワード再設定ページ。URL の ?token と新パスワードを検証 API に渡す。
// 未ログインで到達するため、公開ルート(/reset-password/confirm)に置く。
export function PasswordResetConfirmPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token') ?? ''
  const navigate = useNavigate()

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<PasswordResetConfirmValues>({
    resolver: zodResolver(passwordResetConfirmSchema),
    defaultValues: { password: '', confirmPassword: '' },
    mode: 'onBlur',
  })

  const confirmReset = useMutation({
    ...confirmPasswordResetMutation(),
    onSuccess: () => {
      toast.success('パスワードを再設定しました。ログインしてください')
      navigate('/')
    },
    onError: (error) =>
      toast.error(
        error.response?.status === 400
          ? 'リンクが無効か、有効期限(1時間)が切れています。再度お試しください'
          : 'パスワードの再設定に失敗しました',
      ),
  })

  const onSubmit = handleSubmit((values) => {
    confirmReset.mutate({ body: { token, password: values.password } })
  })

  return (
    <div className="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
      <div className="flex w-full max-w-sm flex-col gap-6">
        <Link to="/" className="self-center">
          <Logo markClassName="size-7" wordmarkClassName="text-xl" />
        </Link>
        <Card className="w-full max-w-sm">
          <CardHeader>
            <CardTitle>新しいパスワードの設定</CardTitle>
            <CardDescription>
              {token
                ? '新しいパスワードを入力してください。'
                : 'リンクが正しくありません。パスワード再設定をやり直してください。'}
            </CardDescription>
          </CardHeader>
          {token ? (
            <form onSubmit={onSubmit}>
              <CardContent>
                <div className="flex flex-col gap-6">
                  <div className="grid gap-2">
                    <Label htmlFor="password">新しいパスワード</Label>
                    <Controller
                      control={control}
                      name="password"
                      render={({ field }) => (
                        <Input
                          id="password"
                          type="password"
                          value={field.value}
                          onChange={field.onChange}
                          onBlur={field.onBlur}
                        />
                      )}
                    />
                    {errors.password && (
                      <p className="text-destructive text-sm">{errors.password.message}</p>
                    )}
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="confirmPassword">新しいパスワード(確認)</Label>
                    <Controller
                      control={control}
                      name="confirmPassword"
                      render={({ field }) => (
                        <Input
                          id="confirmPassword"
                          type="password"
                          value={field.value}
                          onChange={field.onChange}
                          onBlur={field.onBlur}
                        />
                      )}
                    />
                    {errors.confirmPassword && (
                      <p className="text-destructive text-sm">{errors.confirmPassword.message}</p>
                    )}
                  </div>
                </div>
              </CardContent>
              <CardFooter className="mt-6 flex-col gap-2">
                <Button type="submit" className="w-full" disabled={confirmReset.isPending}>
                  パスワードを設定
                </Button>
              </CardFooter>
            </form>
          ) : (
            <CardContent>
              <Button asChild variant="outline" className="w-full">
                <Link to="/reset-password">パスワード再設定をやり直す</Link>
              </Button>
            </CardContent>
          )}
        </Card>
      </div>
    </div>
  )
}
