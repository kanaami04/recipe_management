import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Users } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import { MultiSelectInput } from '@/features/recipes/components/MultiSelectInput'
import { RecipeSharedAvatars } from '@/features/recipes/components/RecipeSharedAvatars'
import {
  getShoppingListQueryKey,
  listUsersOptions,
  updateShoppingListSharesMutation,
} from '@/shared/api/generated/@tanstack/react-query.gen'
import type { UserListItem } from '@/shared/api/generated/types.gen'
import { Button } from '@/shared/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/shared/ui/dialog'

type ShoppingListShareDialogProps = {
  listId: string
  sharedUsers: UserListItem[]
}

// 買い物リストの共有相手を編集するダイアログ。トリガーは共有相手のアバターと人数を表示する。
// 共有相手も同じ 1 つのリストを共同編集する(世帯で使い回す想定)。
export function ShoppingListShareDialog({ listId, sharedUsers }: ShoppingListShareDialogProps) {
  const queryClient = useQueryClient()
  const { data: userCandidates } = useQuery(listUsersOptions())

  const [isOpen, setIsOpen] = useState(false)
  const [selected, setSelected] = useState<string[]>(sharedUsers.map((u) => u.username))

  // 開く瞬間だけ現在の共有相手で選択を初期化する。開いている間の再取得(フォーカス復帰など)で
  // sharedUsers の参照が変わっても、編集中の選択を上書きしない。
  const handleOpenChange = (open: boolean) => {
    if (open) setSelected(sharedUsers.map((u) => u.username))
    setIsOpen(open)
  }

  const updateShares = useMutation({
    ...updateShoppingListSharesMutation(),
    onSuccess: (data) => {
      queryClient.setQueryData(getShoppingListQueryKey(), data)
      setIsOpen(false)
      toast.success('共有相手を更新しました')
    },
    onError: () => toast.error('共有相手の更新に失敗しました'),
  })

  const handleSave = () => {
    updateShares.mutate({
      path: { id: listId },
      body: { shared_user: selected.map((username) => ({ username })) },
    })
  }

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2" aria-label="共有相手を編集">
          <Users className="size-4" />
          {sharedUsers.length > 0 ? <RecipeSharedAvatars users={sharedUsers} /> : <span>共有</span>}
        </Button>
      </DialogTrigger>
      <DialogContent className="flex w-full flex-col sm:max-w-md">
        <DialogHeader>
          <DialogTitle>買い物リストを共有</DialogTitle>
          <DialogDescription>
            共有した相手も同じリストを編集できます。チェックや追加はお互いに反映されます。
          </DialogDescription>
        </DialogHeader>
        <MultiSelectInput
          label="共有相手"
          value={selected}
          placeholder="共有する相手を選択"
          onChange={setSelected}
          options={(userCandidates ?? []).map((u) => ({ label: u.username, value: u.username }))}
        />
        <DialogFooter>
          <Button onClick={handleSave} disabled={updateShares.isPending}>
            保存
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
