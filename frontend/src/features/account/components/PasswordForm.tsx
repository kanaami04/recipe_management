import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation } from '@tanstack/react-query'
import { Controller, useForm } from 'react-hook-form'
import { toast } from 'sonner'

import {
  passwordFormSchema,
  type PasswordFormValues,
} from '@/features/account/schema/accountSchema'
import { changePasswordMutation } from '@/shared/api/generated/@tanstack/react-query.gen'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'

// パスワード変更フォーム。現在のパスワード確認 + 新パスワード(確認入力あり)。
export function PasswordForm() {
  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<PasswordFormValues>({
    resolver: zodResolver(passwordFormSchema),
    defaultValues: { currentPassword: '', newPassword: '', confirmPassword: '' },
    mode: 'onBlur',
  })

  const changePassword = useMutation({
    ...changePasswordMutation(),
    onSuccess: () => {
      reset()
      toast.success('パスワードを変更しました')
    },
    onError: (error) =>
      toast.error(
        error.response?.status === 400
          ? '現在のパスワードが違います'
          : 'パスワードの変更に失敗しました',
      ),
  })

  const onSubmit = handleSubmit((values) => {
    changePassword.mutate({
      body: { current_password: values.currentPassword, new_password: values.newPassword },
    })
  })

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-4">
      <div className="grid gap-2">
        <Label htmlFor="currentPassword">現在のパスワード</Label>
        <Controller
          control={control}
          name="currentPassword"
          render={({ field }) => (
            <Input
              id="currentPassword"
              type="password"
              value={field.value}
              onChange={field.onChange}
              onBlur={field.onBlur}
            />
          )}
        />
        {errors.currentPassword && (
          <p className="text-destructive text-sm">{errors.currentPassword.message}</p>
        )}
      </div>
      <div className="grid gap-2">
        <Label htmlFor="newPassword">新しいパスワード</Label>
        <Controller
          control={control}
          name="newPassword"
          render={({ field }) => (
            <Input
              id="newPassword"
              type="password"
              value={field.value}
              onChange={field.onChange}
              onBlur={field.onBlur}
            />
          )}
        />
        {errors.newPassword && (
          <p className="text-destructive text-sm">{errors.newPassword.message}</p>
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
      <Button type="submit" className="self-start" disabled={changePassword.isPending}>
        パスワードを変更
      </Button>
    </form>
  )
}
