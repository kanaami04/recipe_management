import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { Controller, useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'

import { registerMutation } from '@/shared/api/generated/@tanstack/react-query.gen'
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

import { signUpFormSchema, type SignUpFormValues } from '../schema/signUpFormSchema'

export function SignUpForm() {
  const navigate = useNavigate()

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<SignUpFormValues>({
    resolver: zodResolver(signUpFormSchema),
    defaultValues: { username: '', email: '', password: '' },
    mode: 'onBlur',
  })

  // 登録は生成 mutation。確認メールを送るので、自動ログインはせずログイン画面へ誘導する
  //(確認が済むまでログインできない)。
  const register = useMutation({
    ...registerMutation(),
    onSuccess: () => {
      toast.success('確認メールを送信しました。メール内のリンクから確認を完了してください')
      navigate('/')
    },
    onError: () =>
      toast.error('登録に失敗しました。ユーザー名またはメールが既に使われている可能性があります'),
  })

  const onSubmit = handleSubmit((values) => {
    register.mutate({ body: values })
  })

  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <CardTitle>アカウント作成</CardTitle>
        <CardDescription>アカウントを作成してレシピを管理しましょう</CardDescription>
        <CardAction>
          <Button variant="link" onClick={() => navigate('/')}>
            ログイン
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
              {errors.email && <p className="text-destructive text-sm">{errors.email.message}</p>}
            </div>
            <div className="grid gap-2">
              <Label htmlFor="password">パスワード</Label>
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
          <Button type="submit" className="w-full" disabled={register.isPending}>
            登録
          </Button>
        </CardFooter>
      </form>
    </Card>
  )
}
