import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'

import { loginFormSchema, type LoginFormValues } from '@/features/auth/schema/loginFormSchema'
import { login } from '@/shared/auth/authClient'
import { MessageAlertDialog } from '@/shared/components/MessageAlertDialog'
import { Button } from '@/shared/ui/button'
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'

export function LoginForm() {
  const navigate = useNavigate()
  const [isErrorOpen, setIsErrorOpen] = useState(false)

  // フォーム状態は RHF + zod で管理する (ADR-0006)。
  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginFormSchema),
    defaultValues: { username: '', password: '' },
    mode: 'onBlur',
  })

  const onSubmit = handleSubmit(async (values) => {
    try {
      // access はメモリ保持、refresh は httpOnly Cookie で発行される (ADR-0004)。
      await login(values.username, values.password)
      navigate('/top')
    } catch (error) {
      console.error(error)
      setIsErrorOpen(true)
    }
  })

  return (
    <>
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>アカウントにログイン</CardTitle>
          <CardDescription>ユーザー名とパスワードを入力してください</CardDescription>
          <CardAction>
            <Button variant="link" onClick={() => navigate('/signup')}>
              新規登録
            </Button>
          </CardAction>
        </CardHeader>
        <form onSubmit={onSubmit}>
          <CardContent>
            <div className="flex flex-col gap-6">
              <div className="grid gap-2">
                <Label htmlFor="username">ユーザー名</Label>
                <Controller
                  control={control}
                  name="username"
                  render={({ field }) => (
                    <Input
                      id="username"
                      type="text"
                      value={field.value}
                      onChange={field.onChange}
                      onBlur={field.onBlur}
                    />
                  )}
                />
                {errors.username && (
                  <p className="text-destructive text-sm">{errors.username.message}</p>
                )}
              </div>
              <div className="grid gap-2">
                <div className="flex items-center">
                  <Label htmlFor="password">パスワード</Label>
                  <a
                    href="#"
                    className="ml-auto inline-block text-sm underline-offset-4 hover:underline"
                  >
                    パスワードをお忘れですか？
                  </a>
                </div>
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
            </div>
          </CardContent>
          <CardFooter className="mt-6 flex-col gap-2">
            <Button type="submit" className="w-full">
              ログイン
            </Button>
          </CardFooter>
        </form>
      </Card>

      <MessageAlertDialog
        title="認証に失敗しました"
        description={`ユーザー名またはパスワードが間違っています。\nもう一度入力してください。`}
        open={isErrorOpen}
        onOpenChange={() => setIsErrorOpen(false)}
      />
    </>
  )
}
