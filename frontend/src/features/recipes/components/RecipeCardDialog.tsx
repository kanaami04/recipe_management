import { DialogDescription } from '@radix-ui/react-dialog'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Archive, ArchiveRestore, MoreVertical, Pencil, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import {
  archiveRecipeMutation,
  deleteRecipeMutation,
  listRecipesQueryKey,
} from '@/shared/api/generated/@tanstack/react-query.gen'
import type { RecipeResponse } from '@/shared/api/generated/types.gen'
import { ConfirmDialog } from '@/shared/components/ConfirmDialog'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/shared/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/shared/ui/dropdown-menu'

import { RecipeCard } from './RecipeCard'
import { RecipeDetail } from './RecipeDetail'
import { RecipeDetailEditDialog } from './RecipeDetailEditDialog'

export function RecipeCardDialog({ recipe }: { recipe: RecipeResponse }) {
  const [isEditing, setIsEditing] = useState(false)
  const [isOpen, setIsOpen] = useState(false)
  const [isConfirmingDelete, setIsConfirmingDelete] = useState(false)
  const queryClient = useQueryClient()

  // 削除は生成 mutation + 一覧 query の無効化に集約する。
  const deleteMutation = useMutation({
    ...deleteRecipeMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: listRecipesQueryKey() })
      toast.success('レシピを削除しました')
      setIsOpen(false)
    },
    onError: () => toast.error('削除に失敗しました'),
  })

  const handleDeleteRecipe = () => {
    setIsConfirmingDelete(false)
    deleteMutation.mutate({ path: { id: recipe.id } })
  }

  // アーカイブはユーザーごとの状態。専用 API で自分の状態だけ切り替える
  // (共有相手の状態には影響しない)。
  const archiveMutation = useMutation({
    ...archiveRecipeMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: listRecipesQueryKey() })
      toast.success(recipe.archive_flg ? 'アーカイブを解除しました' : 'レシピをアーカイブしました')
      setIsOpen(false)
    },
    onError: () => toast.error('アーカイブの更新に失敗しました'),
  })

  const handleToggleArchive = () => {
    archiveMutation.mutate({
      path: { id: recipe.id },
      body: { archive_flg: !recipe.archive_flg },
    })
  }

  return (
    <>
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogTrigger asChild>
          <button type="button" className="block h-full w-full cursor-pointer text-left">
            <RecipeCard recipe={recipe} />
          </button>
        </DialogTrigger>
        <DialogContent className="flex max-h-[90dvh] w-full flex-col sm:max-w-3xl">
          <DialogHeader className="pr-16">
            <DialogTitle>{recipe.title}</DialogTitle>
            <DialogDescription>レシピの詳細</DialogDescription>
          </DialogHeader>
          {/* 編集・削除・アーカイブは右上の ⋮ メニューに集約する。閉じるは DialogContent 既定の×。
              modal={false}: メニュー選択で詳細ダイアログを閉じる操作(アーカイブ)時に、
              Radix の body pointer-events ロックが残るのを避ける。 */}
          <DropdownMenu modal={false}>
            <DropdownMenuTrigger asChild>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                aria-label="操作メニュー"
                className="absolute top-3 right-12 size-8"
              >
                <MoreVertical />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => setIsEditing(true)}>
                <Pencil />
                編集
              </DropdownMenuItem>
              <DropdownMenuItem disabled={archiveMutation.isPending} onClick={handleToggleArchive}>
                {recipe.archive_flg ? <ArchiveRestore /> : <Archive />}
                {recipe.archive_flg ? 'アーカイブを解除' : 'アーカイブする'}
              </DropdownMenuItem>
              <DropdownMenuItem
                variant="destructive"
                disabled={deleteMutation.isPending}
                onClick={() => setIsConfirmingDelete(true)}
              >
                <Trash2 />
                削除
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
          <div className="scrollbar-none flex-1 overflow-auto pr-1">
            <RecipeDetail recipe={recipe} />
          </div>
        </DialogContent>
        <RecipeDetailEditDialog
          recipe={recipe}
          open={isEditing}
          onOpenChange={() => setIsEditing(false)}
        />
      </Dialog>
      <ConfirmDialog
        title="レシピを削除しますか？"
        description={`「${recipe.title}」を削除します。\nこの操作は取り消せません。`}
        open={isConfirmingDelete}
        onOpenChange={setIsConfirmingDelete}
        onConfirm={handleDeleteRecipe}
        confirmLabel="削除"
        destructive
      />
    </>
  )
}
