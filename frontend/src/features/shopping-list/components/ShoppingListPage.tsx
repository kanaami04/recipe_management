import {
  closestCenter,
  DndContext,
  type DragEndEvent,
  MouseSensor,
  TouchSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import { arrayMove, SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Trash2 } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import { RecipeSharedAvatars } from '@/features/recipes/components/RecipeSharedAvatars'
import { SortableShoppingListItem } from '@/features/shopping-list/components/SortableShoppingListItem'
import {
  addShoppingListItemMutation,
  clearCheckedShoppingListItemsMutation,
  deleteShoppingListItemMutation,
  getShoppingListOptions,
  getShoppingListQueryKey,
  reorderShoppingListItemsMutation,
  updateShoppingListItemMutation,
} from '@/shared/api/generated/@tanstack/react-query.gen'
import type {
  ShoppingListItemResponse,
  ShoppingListResponse,
} from '@/shared/api/generated/types.gen'
import { ConfirmDialog } from '@/shared/components/ConfirmDialog'
import { Button } from '@/shared/ui/button'
import { Checkbox } from '@/shared/ui/checkbox'
import { Input } from '@/shared/ui/input'
import { Separator } from '@/shared/ui/separator'
import { SidebarTrigger } from '@/shared/ui/sidebar'

// 買い物リスト画面。1 ユーザー 1 リストを使い回す。テキストで品物を追加し、買ったら
// チェックして下部へ沈め、会計後に「チェック済みを削除」でまとめて片付ける。共有相手も
// 同じリストを共同編集する。
export function ShoppingListPage() {
  const queryClient = useQueryClient()
  const { data: list, isPending, isError } = useQuery(getShoppingListOptions())

  const [newName, setNewName] = useState('')
  const [confirmClearOpen, setConfirmClearOpen] = useState(false)

  // サーバは更新後のリスト全体を返すので、キャッシュを差し替えて即時反映する。
  const setList = (data: ShoppingListResponse) =>
    queryClient.setQueryData(getShoppingListQueryKey(), data)

  const addItem = useMutation({
    ...addShoppingListItemMutation(),
    onSuccess: (data) => {
      setList(data)
      setNewName('')
    },
    onError: () => toast.error('品物の追加に失敗しました'),
  })

  // チェックの切り替えは頻繁なので楽観更新で即座に反応させ、確定後はサーバの並びで置き換える。
  const toggleItem = useMutation({
    ...updateShoppingListItemMutation(),
    onMutate: async (vars) => {
      const queryKey = getShoppingListQueryKey()
      await queryClient.cancelQueries({ queryKey })
      const previous = queryClient.getQueryData<ShoppingListResponse>(queryKey)
      queryClient.setQueryData<ShoppingListResponse>(queryKey, (old) =>
        old
          ? {
              ...old,
              items: old.items.map((it) =>
                it.id === vars.path?.item_id ? { ...it, checked: vars.body.checked } : it,
              ),
            }
          : old,
      )
      return { previous }
    },
    onError: (_err, _vars, context) => {
      if (context?.previous) setList(context.previous)
      toast.error('チェックの更新に失敗しました')
    },
    onSuccess: (data) => setList(data),
  })

  const deleteItem = useMutation({
    ...deleteShoppingListItemMutation(),
    onSuccess: (data) => setList(data),
    onError: () => toast.error('品物の削除に失敗しました'),
  })

  const clearChecked = useMutation({
    ...clearCheckedShoppingListItemsMutation(),
    onSuccess: (data) => {
      setList(data)
      setConfirmClearOpen(false)
      toast.success('チェック済みを削除しました')
    },
    onError: () => toast.error('チェック済みの削除に失敗しました'),
  })

  // 並び替えは楽観更新で即座に反映し、確定後はサーバの並びで置き換える(チェックと同じ流儀)。
  const reorder = useMutation({
    ...reorderShoppingListItemsMutation(),
    onMutate: async (vars) => {
      const queryKey = getShoppingListQueryKey()
      await queryClient.cancelQueries({ queryKey })
      const previous = queryClient.getQueryData<ShoppingListResponse>(queryKey)
      if (previous) {
        const byId = new Map(previous.items.map((it) => [it.id, it]))
        const items = vars.body.item_ids
          .map((id) => byId.get(id))
          .filter((it): it is ShoppingListItemResponse => it != null)
        queryClient.setQueryData<ShoppingListResponse>(queryKey, { ...previous, items })
      }
      return { previous }
    },
    onError: (_err, _vars, context) => {
      if (context?.previous) setList(context.previous)
      toast.error('並び替えの保存に失敗しました')
    },
    onSuccess: (data) => setList(data),
  })

  const handleAdd = () => {
    const name = newName.trim()
    // Enter 連打での二重送信を防ぐ(newName は onSuccess でしか消えないため isPending も見る)。
    if (name === '' || !list || addItem.isPending) return
    addItem.mutate({ path: { id: list.id }, body: { name } })
  }

  // 未チェックだけを並び替え対象にする。チェック済み(購入済み)は末尾に固定表示する。
  // サーバは checked → position → id 順で返すので、filter で分けると各グループの順序を保てる。
  const items = list?.items ?? []
  const uncheckedItems = items.filter((it) => !it.checked)
  const checkedItems = items.filter((it) => it.checked)

  // マウスは 8px 動いて初めてドラッグ開始、タッチは 200ms の長押しで開始(短押しはタップ)。
  const sensors = useSensors(
    useSensor(MouseSensor, { activationConstraint: { distance: 8 } }),
    useSensor(TouchSensor, { activationConstraint: { delay: 200, tolerance: 8 } }),
  )

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id || !list) return
    const oldIndex = uncheckedItems.findIndex((it) => it.id === active.id)
    const newIndex = uncheckedItems.findIndex((it) => it.id === over.id)
    if (oldIndex < 0 || newIndex < 0) return
    // 並び替え後の未チェック + 末尾のチェック済み、を全体の順序としてサーバへ送る。
    const nextUnchecked = arrayMove(uncheckedItems, oldIndex, newIndex)
    const itemIds = [...nextUnchecked, ...checkedItems].map((it) => it.id)
    reorder.mutate({ path: { id: list.id }, body: { item_ids: itemIds } })
  }

  return (
    <>
      <header className="flex h-(--header-height) shrink-0 items-center gap-2 sticky top-0 z-10 border-b bg-background transition-[width,height] ease-linear">
        <div className="flex w-full items-center gap-2 px-3 sm:px-4 lg:px-6">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mx-1 data-[orientation=vertical]:h-4" />
          <h1 className="text-base font-medium">買い物リスト</h1>
          {/* 共有中はグループメンバーのアバターを表示(共有の管理は共有グループ画面で行う)。 */}
          {list && list.shared_user.length > 0 && (
            <div className="ml-auto">
              <RecipeSharedAvatars users={list.shared_user} />
            </div>
          )}
        </div>
      </header>

      <div className="mx-auto flex w-full max-w-2xl flex-col gap-4 p-3 sm:p-4">
        {/* 追加行。Enter で連続追加できる。 */}
        <div className="flex gap-2">
          <Input
            placeholder="追加する品物"
            value={newName}
            maxLength={50}
            disabled={!list}
            onChange={(e) => setNewName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleAdd()
            }}
          />
          <Button
            onClick={handleAdd}
            disabled={addItem.isPending || !list || newName.trim() === ''}
          >
            追加
          </Button>
        </div>

        {isPending ? (
          <p className="text-muted-foreground py-8 text-center">読み込み中...</p>
        ) : isError ? (
          <p className="text-destructive py-8 text-center">買い物リストの取得に失敗しました</p>
        ) : items.length === 0 ? (
          <p className="text-muted-foreground py-8 text-center">品物がまだありません</p>
        ) : (
          <DndContext
            sensors={sensors}
            collisionDetection={closestCenter}
            onDragEnd={handleDragEnd}
          >
            {/* 先頭・末尾の行(bg-card)の角をコンテナの丸角に合わせて丸め、四隅が被さらないようにする */}
            <ul className="divide-border divide-y rounded-md border [&>li:first-child]:rounded-t-md [&>li:last-child]:rounded-b-md">
              {/* 未チェックはグリップの長押し/ドラッグで並び替え可能 */}
              <SortableContext
                items={uncheckedItems.map((it) => it.id)}
                strategy={verticalListSortingStrategy}
              >
                {uncheckedItems.map((item) => (
                  <SortableShoppingListItem
                    key={item.id}
                    item={item}
                    onToggle={(checked) =>
                      toggleItem.mutate({
                        path: { id: list!.id, item_id: item.id },
                        body: { checked },
                      })
                    }
                    onDelete={() => deleteItem.mutate({ path: { id: list!.id, item_id: item.id } })}
                  />
                ))}
              </SortableContext>
              {/* チェック済み(購入済み)は末尾に固定。並び替え対象外なのでグリップは出さない */}
              {checkedItems.map((item) => (
                <li key={item.id} className="flex items-center gap-3 p-3 pl-10">
                  <label className="flex flex-1 cursor-pointer items-center gap-3">
                    <Checkbox
                      checked={item.checked}
                      onCheckedChange={(checked) =>
                        toggleItem.mutate({
                          path: { id: list!.id, item_id: item.id },
                          body: { checked: checked === true },
                        })
                      }
                    />
                    <span className="text-muted-foreground truncate line-through">{item.name}</span>
                  </label>
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label="削除"
                    onClick={() => deleteItem.mutate({ path: { id: list!.id, item_id: item.id } })}
                  >
                    <Trash2 />
                  </Button>
                </li>
              ))}
            </ul>
          </DndContext>
        )}

        {checkedItems.length > 0 && list && (
          <div className="flex justify-end">
            <Button
              variant="outline"
              onClick={() => setConfirmClearOpen(true)}
              disabled={clearChecked.isPending}
            >
              チェック済みを削除({checkedItems.length})
            </Button>
          </div>
        )}
      </div>

      <ConfirmDialog
        title="チェック済みを削除しますか？"
        description={`チェック済みの ${checkedItems.length} 件を削除します。`}
        open={confirmClearOpen}
        onOpenChange={setConfirmClearOpen}
        onConfirm={() => list && clearChecked.mutate({ path: { id: list.id } })}
        confirmLabel="削除"
        destructive
      />
    </>
  )
}
