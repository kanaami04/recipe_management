import { ArrowLeft } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

import { EmailForm } from '@/features/account/components/EmailForm'
import { Button } from '@/shared/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { Separator } from '@/shared/ui/separator'
import { SidebarTrigger } from '@/shared/ui/sidebar'

// メールアドレス変更専用画面。本人確認(現在のパスワード)のうえ変更し、
// 成功したら再ログインを求める(EmailForm 側でログアウト + 画面遷移)。
export function EmailChangePage() {
  const navigate = useNavigate()

  return (
    <>
      <header className="flex h-(--header-height) shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear">
        <div className="flex w-full items-center gap-2 px-3 sm:px-4 lg:px-6">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mx-1 data-[orientation=vertical]:h-4" />
          <h1 className="text-base font-medium">メールアドレス変更</h1>
        </div>
      </header>

      <div className="mx-auto flex w-full max-w-2xl flex-col gap-4 p-3 sm:p-4">
        <Button
          type="button"
          variant="ghost"
          className="w-fit"
          onClick={() => navigate('/top/account')}
        >
          <ArrowLeft />
          アカウントに戻る
        </Button>

        <Card>
          <CardHeader>
            <CardTitle>メールアドレス変更</CardTitle>
            <p className="text-muted-foreground text-sm">
              変更すると自動的にログアウトします。新しいメールアドレスで再度ログインしてください。
            </p>
          </CardHeader>
          <CardContent>
            <EmailForm />
          </CardContent>
        </Card>
      </div>
    </>
  )
}
