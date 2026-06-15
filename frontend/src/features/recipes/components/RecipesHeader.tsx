import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { toast } from 'sonner'

import { RecipeForm } from '@/features/recipes/components/RecipeForm'
import {
  createRecipeMutation,
  listLabelsOptions,
  listRecipesQueryKey,
  listUsersOptions,
} from '@/shared/api/generated/@tanstack/react-query.gen'
import type { RecipeRequest } from '@/shared/api/generated/types.gen'
import { Button } from '@/shared/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Separator } from '@/shared/ui/separator'
import { SidebarTrigger } from '@/shared/ui/sidebar'

export function RecipesHeader() {
  const queryClient = useQueryClient()
  const { data: sharedUserData } = useQuery(listUsersOptions())
  const { data: labelData } = useQuery(listLabelsOptions())

  const [isOpen, setIsOpen] = useState(false)

  // 作成は生成 mutation + 一覧 query の無効化に集約する (ADR-0003)。
  const createMutation = useMutation({
    ...createRecipeMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: listRecipesQueryKey() })
      toast.success('レシピを作成しました')
      setIsOpen(false)
    },
    onError: () => toast.error('レシピの作成に失敗しました'),
  })

  const handleCreate = async (payload: RecipeRequest) => {
    createMutation.mutate({ body: payload })
  }

  return (
    <header className="flex h-(--header-height) shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-(--header-height)">
      <div className="flex w-full items-center gap-1 lg:gap-2 lg:px-6">
        <SidebarTrigger className="-ml-1" />
        <Separator orientation="vertical" className="mx-1 data-[orientation=vertical]:h-4" />
        <h1 className="text-base font-medium">MyRecipes</h1>
        <div className="ml-auto flex items-center gap-2">
          <Input placeholder="search" className="sm:w-40" />
          <Separator orientation="vertical" className="data-[orientation=vertical]:h-4" />
          <Dialog open={isOpen} onOpenChange={setIsOpen}>
            <DialogTrigger asChild>
              <Button size="sm" onClick={() => setIsOpen(true)}>
                {' '}
                +{' '}
              </Button>
            </DialogTrigger>
            <DialogContent className="flex max-h-[90dvh] w-full flex-col sm:max-w-3xl">
              <DialogHeader>
                <DialogTitle>new recipe</DialogTitle>
                <DialogDescription>create new recipe.</DialogDescription>
              </DialogHeader>
              <RecipeForm
                mode="create"
                onSubmit={handleCreate}
                labelData={labelData}
                sharedUserData={sharedUserData}
              />
            </DialogContent>
          </Dialog>
        </div>
      </div>
    </header>
  )
}
