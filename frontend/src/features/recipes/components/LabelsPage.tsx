import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import { Check, Pencil, Trash2, X } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import {
  createLabelMutation,
  deleteLabelMutation,
  listLabelsOptions,
  listLabelsQueryKey,
  listRecipesQueryKey,
  updateLabelMutation,
} from '@/shared/api/generated/@tanstack/react-query.gen'
import type { LabelItem } from '@/shared/api/generated/types.gen'
import { ConfirmDialog } from '@/shared/components/ConfirmDialog'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Separator } from '@/shared/ui/separator'
import { SidebarTrigger } from '@/shared/ui/sidebar'

// 409(重複)なら重複メッセージ、それ以外は fallback を返す。
const dupOrMessage = (error: AxiosError, fallback: string) =>
  error.response?.status === 409 ? '同名のラベルが既にあります' : fallback

// ラベル管理画面。自分のラベルの作成・改名・削除を行う。改名・削除は自分のレシピにも伝播する。
export function LabelsPage() {
  const queryClient = useQueryClient()
  const { data: labels, isPending, isError } = useQuery(listLabelsOptions())

  const [newName, setNewName] = useState('')
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editingName, setEditingName] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<LabelItem | null>(null)

  const invalidateLabels = () => queryClient.invalidateQueries({ queryKey: listLabelsQueryKey() })
  // 改名・削除は recipe_labels にも伝播するため、レシピ一覧のキャッシュも無効化する。
  const invalidateLabelsAndRecipes = () => {
    invalidateLabels()
    queryClient.invalidateQueries({ queryKey: listRecipesQueryKey() })
  }

  const createLabel = useMutation({
    ...createLabelMutation(),
    onSuccess: () => {
      invalidateLabels()
      setNewName('')
      toast.success('ラベルを追加しました')
    },
    onError: (error) => toast.error(dupOrMessage(error, 'ラベルの追加に失敗しました')),
  })

  const renameLabel = useMutation({
    ...updateLabelMutation(),
    onSuccess: () => {
      invalidateLabelsAndRecipes()
      setEditingId(null)
      toast.success('ラベル名を変更しました')
    },
    onError: (error) => toast.error(dupOrMessage(error, 'ラベル名の変更に失敗しました')),
  })

  const removeLabel = useMutation({
    ...deleteLabelMutation(),
    onSuccess: () => {
      invalidateLabelsAndRecipes()
      setDeleteTarget(null)
      toast.success('ラベルを削除しました')
    },
    onError: () => toast.error('ラベルの削除に失敗しました'),
  })

  const handleCreate = () => {
    const name = newName.trim()
    if (name === '') return
    createLabel.mutate({ body: { name } })
  }

  const startEdit = (label: LabelItem) => {
    setEditingId(label.id)
    setEditingName(label.name)
  }

  const handleRename = () => {
    const name = editingName.trim()
    if (name === '' || editingId === null) return
    renameLabel.mutate({ path: { id: editingId }, body: { name } })
  }

  return (
    <>
      <header className="flex h-(--header-height) shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear">
        <div className="flex w-full items-center gap-2 px-3 sm:px-4 lg:px-6">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mx-1 data-[orientation=vertical]:h-4" />
          <h1 className="text-base font-medium">ラベル管理</h1>
        </div>
      </header>

      <div className="mx-auto flex w-full max-w-2xl flex-col gap-4 p-3 sm:p-4">
        {/* 新規作成 */}
        <div className="flex gap-2">
          <Input
            placeholder="新しいラベル名"
            value={newName}
            maxLength={50}
            onChange={(e) => setNewName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleCreate()
            }}
          />
          <Button onClick={handleCreate} disabled={createLabel.isPending || newName.trim() === ''}>
            追加
          </Button>
        </div>

        {isPending ? (
          <p className="text-muted-foreground py-8 text-center">読み込み中...</p>
        ) : isError ? (
          <p className="text-destructive py-8 text-center">ラベルの取得に失敗しました</p>
        ) : labels.length === 0 ? (
          <p className="text-muted-foreground py-8 text-center">ラベルがまだありません</p>
        ) : (
          <ul className="divide-border divide-y rounded-md border">
            {labels.map((label) => (
              <li key={label.id} className="flex items-center gap-2 p-2">
                {editingId === label.id ? (
                  <>
                    <Input
                      value={editingName}
                      maxLength={50}
                      autoFocus
                      onChange={(e) => setEditingName(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') handleRename()
                        if (e.key === 'Escape') setEditingId(null)
                      }}
                    />
                    <Button
                      variant="outline"
                      size="icon"
                      aria-label="保存"
                      disabled={renameLabel.isPending || editingName.trim() === ''}
                      onClick={handleRename}
                    >
                      <Check />
                    </Button>
                    <Button
                      variant="outline"
                      size="icon"
                      aria-label="キャンセル"
                      onClick={() => setEditingId(null)}
                    >
                      <X />
                    </Button>
                  </>
                ) : (
                  <>
                    <span className="flex-1 truncate">{label.name}</span>
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label="名前を変更"
                      onClick={() => startEdit(label)}
                    >
                      <Pencil />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label="削除"
                      onClick={() => setDeleteTarget(label)}
                    >
                      <Trash2 />
                    </Button>
                  </>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>

      <ConfirmDialog
        title="ラベルを削除しますか？"
        description={
          deleteTarget
            ? `「${deleteTarget.name}」を削除します。\n自分のレシピからもこのラベルが外れます。`
            : ''
        }
        open={deleteTarget !== null}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        onConfirm={() => deleteTarget && removeLabel.mutate({ path: { id: deleteTarget.id } })}
        confirmLabel="削除"
        destructive
      />
    </>
  )
}
