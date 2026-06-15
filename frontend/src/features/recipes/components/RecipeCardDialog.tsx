import { DialogDescription } from '@radix-ui/react-dialog'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { toast } from 'sonner'

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

  // 削除は生成 mutation + 一覧 query の無効化に集約する (ADR-0003)。
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
          <div role="button" style={{ display: 'inline-block', cursor: 'pointer' }}>
            <RecipeCard key={recipe.id} recipe={recipe} />
          </div>
        </DialogTrigger>
        <DialogContent className="max-w-3xl w-full">
          <DialogHeader>
            <DialogTitle>{recipe.title}</DialogTitle>
            <DialogDescription>レシピの詳細</DialogDescription>
          </DialogHeader>
          <div className="gap-4 h-140 overflow-auto">
            <div className="grid gap-3">
              <div className="flex gap-3 w-full justify-start">
                <Label>{recipe.create_for}人前</Label>
                <Label>{recipe.create_time}分</Label>
              </div>
              <div className="flex-2 gap-3">
                <Label>食材</Label>
                {recipe.cooking.map((c) => (
                  <p key={c.ingredients.name}>
                    {c.ingredients.name} : {c.quantity}
                    {c.unit}
                  </p>
                ))}

                <Label>調味料</Label>
                {recipe.season.map((s) => (
                  <p key={s.seasoning.name}>
                    {s.seasoning.name} : {s.quantity}
                    {s.unit}
                  </p>
                ))}
                <Label>説明</Label>
                <p>{recipe.procedure}</p>
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
              <Button variant="outline">Cancel</Button>
            </DialogClose>
            <Button type="button" onClick={() => setIsEditing(true)}>
              Edit
            </Button>
            <Button
              type="button"
              variant="destructive"
              disabled={deleteMutation.isPending}
              onClick={() => setIsConfirmingDelete(true)}
            >
              Delete
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
