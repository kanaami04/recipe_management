import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'

import { EmailForm } from '@/features/account/components/EmailForm'
import { PasswordForm } from '@/features/account/components/PasswordForm'
import { ProfileForm } from '@/features/account/components/ProfileForm'
import { deleteAccountMutation } from '@/shared/api/generated/@tanstack/react-query.gen'
import { logout } from '@/shared/auth/authClient'
import { useUser } from '@/shared/auth/UserContext'
import { ConfirmDialog } from '@/shared/components/ConfirmDialog'
import { Button } from '@/shared/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { Separator } from '@/shared/ui/separator'
import { SidebarTrigger } from '@/shared/ui/sidebar'

// アカウント画面。プロフィールの閲覧・編集、パスワード変更、アカウント削除を行う。
export function AccountPage() {
  const { user } = useUser()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [isConfirmingDelete, setIsConfirmingDelete] = useState(false)

  const deleteAccount = useMutation({
    ...deleteAccountMutation(),
    onSuccess: async () => {
      // アカウントが消えたのでログアウト(Cookie 失効 + token 破棄)してログインへ。
      await logout()
      // 別アカウントで再ログインしたとき、削除済みユーザーの情報が一瞬残らないよう
      // キャッシュを空にする(getUserInfo などのユーザー固有データを破棄)。
      queryClient.clear()
      navigate('/')
      toast.success('アカウントを削除しました')
    },
    onError: () => toast.error('アカウントの削除に失敗しました'),
  })

  return (
    <>
      <header className="flex h-(--header-height) shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear">
        <div className="flex w-full items-center gap-2 px-3 sm:px-4 lg:px-6">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mx-1 data-[orientation=vertical]:h-4" />
          <h1 className="text-base font-medium">アカウント</h1>
        </div>
      </header>

      {!user ? (
        <p className="text-muted-foreground py-8 text-center">読み込み中...</p>
      ) : (
        <div className="mx-auto flex w-full max-w-2xl flex-col gap-4 p-3 sm:p-4">
          <Card>
            <CardHeader>
              <CardTitle>プロフィール</CardTitle>
              <p className="text-muted-foreground text-sm">登録日: {user.created_at}</p>
            </CardHeader>
            <CardContent>
              <ProfileForm user={user} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>メールアドレス変更</CardTitle>
            </CardHeader>
            <CardContent>
              <EmailForm />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>パスワード変更</CardTitle>
            </CardHeader>
            <CardContent>
              <PasswordForm />
            </CardContent>
          </Card>

          <Card className="border-destructive/50">
            <CardHeader>
              <CardTitle className="text-destructive">アカウント削除</CardTitle>
              <p className="text-muted-foreground text-sm">
                削除すると、あなたが作成したレシピやラベルもすべて消えます。この操作は取り消せません。
              </p>
            </CardHeader>
            <CardContent>
              <Button
                variant="destructive"
                disabled={deleteAccount.isPending}
                onClick={() => setIsConfirmingDelete(true)}
              >
                アカウントを削除
              </Button>
            </CardContent>
          </Card>
        </div>
      )}

      <ConfirmDialog
        title="アカウントを削除しますか？"
        description={
          '本当にアカウントを削除しますか？\n作成したレシピ・ラベルもすべて削除され、元に戻せません。'
        }
        open={isConfirmingDelete}
        onOpenChange={setIsConfirmingDelete}
        onConfirm={() => {
          setIsConfirmingDelete(false)
          deleteAccount.mutate({})
        }}
        confirmLabel="削除"
        destructive
      />
    </>
  )
}
