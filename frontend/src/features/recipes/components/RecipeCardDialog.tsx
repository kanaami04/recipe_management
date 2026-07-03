import { DialogDescription } from '@radix-ui/react-dialog'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Carrot, Clock, Droplet, ListOrdered, Users } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import { splitSteps } from '@/features/recipes/steps'
import { formatAmount } from '@/features/recipes/units'
import {
  deleteRecipeMutation,
  listRecipesQueryKey,
} from '@/shared/api/generated/@tanstack/react-query.gen'
import type { RecipeResponse } from '@/shared/api/generated/types.gen'
import { ConfirmDialog } from '@/shared/components/ConfirmDialog'
import { Button } from '@/shared/ui/button'
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/shared/ui/dialog'
import { Label } from '@/shared/ui/label'

import { RecipeCard } from './RecipeCard'
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

  return (
    <>
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogTrigger asChild>
          <button type="button" className="block h-full w-full cursor-pointer text-left">
            <RecipeCard recipe={recipe} />
          </button>
        </DialogTrigger>
        <DialogContent className="flex max-h-[90dvh] w-full flex-col sm:max-w-3xl">
          <DialogHeader>
            <DialogTitle>{recipe.title}</DialogTitle>
            <DialogDescription>レシピの詳細</DialogDescription>
          </DialogHeader>
          <div className="flex-1 overflow-auto">
            <div className="grid gap-3">
              <div className="text-muted-foreground flex w-full items-center gap-4 text-sm">
                <span className="flex items-center gap-1">
                  <Users className="size-4" />
                  {recipe.create_for}人前
                </span>
                {recipe.create_time != null && (
                  <span className="flex items-center gap-1">
                    <Clock className="size-4" />
                    {recipe.create_time}分
                  </span>
                )}
              </div>
              <div className="flex-2 gap-3">
                <Label className="flex items-center gap-1">
                  <Carrot className="size-4" />
                  食材
                </Label>
                {recipe.cooking.map((c) => (
                  <p key={c.ingredients.name}>
                    {c.ingredients.name} : {formatAmount(c.quantity, c.unit)}
                  </p>
                ))}

                <Label className="flex items-center gap-1">
                  <Droplet className="size-4" />
                  調味料
                </Label>
                {recipe.season.map((s) => (
                  <p key={s.seasoning.name}>
                    {s.seasoning.name} : {formatAmount(s.quantity, s.unit)}
                  </p>
                ))}
                <Label className="flex items-center gap-1">
                  <ListOrdered className="size-4" />
                  作り方
                </Label>
                <ol className="list-decimal pl-5">
                  {splitSteps(recipe.procedure)
                    .filter((step) => step.trim() !== '')
                    .map((step, index) => (
                      <li key={index}>{step}</li>
                    ))}
                </ol>
              </div>
              <div>
                {recipe.label.map((label) => (
                  <p key={label.name}>{label.name}</p>
                ))}
              </div>
              <div>
                {recipe.shared_user.map((user) => (
                  <p key={user.username}>{user.username}</p>
                ))}
              </div>
            </div>
          </div>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline">閉じる</Button>
            </DialogClose>
            <Button type="button" onClick={() => setIsEditing(true)}>
              編集
            </Button>
            <Button
              type="button"
              variant="destructive"
              disabled={deleteMutation.isPending}
              onClick={() => setIsConfirmingDelete(true)}
            >
              削除
            </Button>
          </DialogFooter>
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
