import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

import {
  listLabelsOptions,
  listRecipesQueryKey,
  listUsersOptions,
  updateRecipeMutation,
} from '@/shared/api/generated/@tanstack/react-query.gen'
import type { RecipeRequest, RecipeResponse } from '@/shared/api/generated/types.gen'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/shared/ui/dialog'

import { RecipeForm } from './RecipeForm'

type EditDialog = {
  recipe: RecipeResponse
  open: boolean
  onOpenChange: () => void
}

export function RecipeDetailEditDialog({ recipe, open, onOpenChange }: EditDialog) {
  const queryClient = useQueryClient()
  const { data: sharedUserData } = useQuery(listUsersOptions())
  const { data: labelData } = useQuery(listLabelsOptions())

  // 更新は生成 mutation + 一覧 query の無効化に集約する (ADR-0003)。
  const updateMutation = useMutation({
    ...updateRecipeMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: listRecipesQueryKey() })
      toast.success('レシピを編集しました')
      onOpenChange()
    },
    onError: () => toast.error('レシピの編集に失敗しました'),
  })

  const handleEdit = async (payload: RecipeRequest) => {
    updateMutation.mutate({ path: { id: recipe.id }, body: payload })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl w-full">
        <DialogHeader>
          <DialogTitle>レシピ編集</DialogTitle>
          <DialogDescription>レシピを編集します</DialogDescription>
        </DialogHeader>
        <RecipeForm
          mode="edit"
          initialData={recipe}
          onSubmit={handleEdit}
          labelData={labelData}
          sharedUserData={sharedUserData}
          onClickCancel={onOpenChange}
        />
      </DialogContent>
    </Dialog>
  )
}
