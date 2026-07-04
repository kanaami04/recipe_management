import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Controller, useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'

import { emailFormSchema, type EmailFormValues } from '@/features/account/schema/accountSchema'
import { changeEmailMutation } from '@/shared/api/generated/@tanstack/react-query.gen'
import { logout } from '@/shared/auth/authClient'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'

// メールアドレス変更フォーム。本人確認のため現在のパスワードを求める。
// メールは新しいログイン識別子になるため、変更が成功したらログアウトして
// ログイン画面へ戻し、新しいメールで再ログインしてもらう。
export function EmailForm() {
  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<EmailFormValues>({
    resolver: zodResolver(emailFormSchema),
    defaultValues: { email: '', password: '' },
    mode: 'onBlur',
  })

  const changeEmail = useMutation({
    ...changeEmailMutation(),
    onSuccess: async () => {
      await logout()
      // 別アカウント(または同アカウントの新メール)で再ログインしたとき、
      // 旧メールの情報が一瞬残らないようキャッシュを空にする。
      queryClient.clear()
      navigate('/')
      toast.success('メールアドレスを変更しました。新しいメールアドレスでログインしてください。')
    },
    onError: (error) => {
      if (error.response?.status === 409) {
        toast.error('そのメールアドレスは既に使われています')
      } else if (error.response?.status === 400) {
        toast.error('パスワードが違います')
      } else {
        toast.error('メールアドレスの変更に失敗しました')
      }
    },
  })

  const onSubmit = handleSubmit((values) => {
    changeEmail.mutate({ body: { email: values.email, password: values.password } })
  })

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-4">
      <div className="grid gap-2">
        <Label htmlFor="newEmail">新しいメールアドレス</Label>
        <Controller
          control={control}
          name="email"
          render={({ field }) => (
            <Input
              id="newEmail"
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
        <Label htmlFor="emailPassword">現在のパスワード</Label>
        <Controller
          control={control}
          name="password"
          render={({ field }) => (
            <Input
              id="emailPassword"
              type="password"
              value={field.value}
              onChange={field.onChange}
              onBlur={field.onBlur}
            />
          )}
        />
        {errors.password && <p className="text-destructive text-sm">{errors.password.message}</p>}
      </div>
      <Button type="submit" className="self-start" disabled={changeEmail.isPending}>
        メールアドレスを変更
      </Button>
    </form>
  )
}
