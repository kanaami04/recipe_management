import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Controller, useForm } from 'react-hook-form'
import { toast } from 'sonner'

import { profileFormSchema, type ProfileFormValues } from '@/features/account/schema/accountSchema'
import {
  getUserInfoQueryKey,
  updateUserInfoMutation,
} from '@/shared/api/generated/@tanstack/react-query.gen'
import type { UserInfoResponse } from '@/shared/api/generated/types.gen'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'

// プロフィール(ユーザー名・メール)の編集フォーム。
export function ProfileForm({ user }: { user: UserInfoResponse }) {
  const queryClient = useQueryClient()
  const {
    control,
    handleSubmit,
    reset,
    formState: { errors, isDirty },
  } = useForm<ProfileFormValues>({
    resolver: zodResolver(profileFormSchema),
    defaultValues: { username: user.username, email: user.email },
    mode: 'onBlur',
  })

  const updateProfile = useMutation({
    ...updateUserInfoMutation(),
    onSuccess: (_data, variables) => {
      // 保存済みの値を新しい既定値にして dirty を解除する(保存ボタンが押しっぱなしにならない)。
      reset(variables.body)
      queryClient.invalidateQueries({ queryKey: getUserInfoQueryKey() })
      toast.success('プロフィールを更新しました')
    },
    onError: (error) =>
      toast.error(
        error.response?.status === 409
          ? 'そのユーザー名またはメールアドレスは既に使われています'
          : 'プロフィールの更新に失敗しました',
      ),
  })

  const onSubmit = handleSubmit((values) => {
    updateProfile.mutate({ body: values })
  })

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-4">
      <div className="grid gap-2">
        <Label htmlFor="username">ユーザー名</Label>
        <Controller
          control={control}
          name="username"
          render={({ field }) => (
            <Input
              id="username"
              value={field.value}
              onChange={field.onChange}
              onBlur={field.onBlur}
            />
          )}
        />
        {errors.username && <p className="text-destructive text-sm">{errors.username.message}</p>}
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
      <Button type="submit" className="self-start" disabled={!isDirty || updateProfile.isPending}>
        保存
      </Button>
    </form>
  )
}
