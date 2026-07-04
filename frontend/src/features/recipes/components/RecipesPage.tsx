import {
  closestCenter,
  DndContext,
  type DragEndEvent,
  MouseSensor,
  TouchSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import { arrayMove, rectSortingStrategy, SortableContext } from '@dnd-kit/sortable'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { type ReactNode, useState } from 'react'
import { toast } from 'sonner'

import { RecipeCardDialog } from '@/features/recipes/components/RecipeCardDialog'
import { RecipesHeader } from '@/features/recipes/components/RecipesHeader'
import { SortableRecipeCard } from '@/features/recipes/components/SortableRecipeCard'
import {
  listRecipesOptions,
  listRecipesQueryKey,
  reorderRecipesMutation,
} from '@/shared/api/generated/@tanstack/react-query.gen'
import type { RecipeResponse } from '@/shared/api/generated/types.gen'

// ラベル絞り込みのチップ。選択中は塗り、未選択は枠線。
function FilterChip({
  active,
  onClick,
  children,
}: {
  active: boolean
  onClick: () => void
  children: ReactNode
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`rounded-full border px-3 py-1 text-sm ${
        active ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'
      }`}
    >
      {children}
    </button>
  )
}

export function RecipesPage() {
  // サーバ状態は生成 Query フックで取得する。認証は interceptor が付与する。
  const { data: allRecipes, isPending, isError } = useQuery(listRecipesOptions())
  const queryClient = useQueryClient()

  // ラベル絞り込み(クライアント側)。null は絞り込みなし。
  const [selectedLabel, setSelectedLabel] = useState<string | null>(null)

  // メイン一覧はアーカイブ済みを除く(アーカイブは /top/archive で表示)。
  const recipesData = (allRecipes ?? []).filter((r) => !r.archive_flg)

  // 絞り込み候補は一覧に実在するラベル名(重複なし・名前順)。
  const availableLabels = [
    ...new Set(recipesData.flatMap((r) => r.label.map((l) => l.name))),
  ].sort()
  // 選択ラベルが候補から消えたら(最後のレシピをアーカイブ/削除した等)絞り込みを解除する。
  // これを実表示に使い、消えたラベルで空表示のまま固まるのを防ぐ。
  const activeLabel =
    selectedLabel !== null && availableLabels.includes(selectedLabel) ? selectedLabel : null
  // 選択中ラベルで絞った表示対象。絞り込み中は並び替えを行わない(部分並び替えを避ける)。
  const displayed =
    activeLabel === null
      ? recipesData
      : recipesData.filter((r) => r.label.some((l) => l.name === activeLabel))

  // 楽観更新の定石: onMutate で進行中の再取得を止めてから並びを先に反映し、
  // 失敗時はスナップショットへ戻す。成否に関わらず onSettled でサーバと再同期する。
  const reorderMutation = useMutation({
    ...reorderRecipesMutation(),
    onMutate: async (vars) => {
      const key = listRecipesQueryKey()
      await queryClient.cancelQueries({ queryKey: key })
      const previous = queryClient.getQueryData<RecipeResponse[]>(key)
      if (previous) {
        const byId = new Map(previous.map((r) => [r.id, r]))
        const reordered = vars.body.recipe_ids
          .map((id) => byId.get(id))
          .filter((r): r is RecipeResponse => r != null)
        // 並び替え対象(非アーカイブ)を新順に、それ以外(アーカイブ等)はそのまま残す。
        const idSet = new Set(vars.body.recipe_ids)
        const next = [...reordered, ...previous.filter((r) => !idSet.has(r.id))]
        queryClient.setQueryData<RecipeResponse[]>(key, next)
      }
      return { previous }
    },
    onError: (_err, _vars, context) => {
      if (context?.previous) {
        queryClient.setQueryData<RecipeResponse[]>(listRecipesQueryKey(), context.previous)
      }
      toast.error('並び替えの保存に失敗しました')
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: listRecipesQueryKey() })
    },
  })

  // マウスは 8px 動いて初めてドラッグ開始(タップ=詳細を開く)。
  // タッチは 200ms の長押しで開始(短いスワイプはスクロール、タップは詳細)。
  const sensors = useSensors(
    useSensor(MouseSensor, { activationConstraint: { distance: 8 } }),
    useSensor(TouchSensor, { activationConstraint: { delay: 200, tolerance: 8 } }),
  )

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id || !recipesData) return
    const oldIndex = recipesData.findIndex((r) => r.id === active.id)
    const newIndex = recipesData.findIndex((r) => r.id === over.id)
    if (oldIndex < 0 || newIndex < 0) return

    const next = arrayMove(recipesData, oldIndex, newIndex)
    reorderMutation.mutate({ body: { recipe_ids: next.map((r) => r.id) } })
  }

  return (
    <>
      <RecipesHeader />
      {isPending ? (
        <div className="flex items-center justify-center min-h-60">
          <p className="text-muted-foreground">読み込み中...</p>
        </div>
      ) : isError ? (
        <div className="flex items-center justify-center min-h-60">
          <p className="text-destructive">レシピの取得に失敗しました</p>
        </div>
      ) : recipesData.length === 0 ? (
        <div className="flex items-center justify-center min-h-60">
          <p className="text-muted-foreground">レシピがまだありません</p>
        </div>
      ) : (
        <>
          {/* ラベル絞り込み。タップで単一ラベルに絞る。「すべて」で解除。 */}
          {availableLabels.length > 0 && (
            <div className="flex flex-wrap gap-2 px-3 pt-3 sm:px-4">
              <FilterChip active={activeLabel === null} onClick={() => setSelectedLabel(null)}>
                すべて
              </FilterChip>
              {availableLabels.map((name) => (
                <FilterChip
                  key={name}
                  active={activeLabel === name}
                  onClick={() => setSelectedLabel(name)}
                >
                  {name}
                </FilterChip>
              ))}
            </div>
          )}
          {/* 絞り込みなしは並び替え可能。絞り込み中はカード直描画(部分並び替えを避ける)。
              activeLabel は実在ラベルのみ(or null)なので、絞り込み結果は必ず1件以上。 */}
          {activeLabel === null ? (
            <DndContext
              sensors={sensors}
              collisionDetection={closestCenter}
              onDragEnd={handleDragEnd}
            >
              <SortableContext items={recipesData.map((r) => r.id)} strategy={rectSortingStrategy}>
                <div className="grid grid-cols-2 gap-3 p-3 sm:grid-cols-3 sm:gap-4 sm:p-4 lg:grid-cols-4 xl:grid-cols-5">
                  {recipesData.map((recipe) => (
                    <SortableRecipeCard key={recipe.id} recipe={recipe} />
                  ))}
                </div>
              </SortableContext>
            </DndContext>
          ) : (
            <div className="grid grid-cols-2 gap-3 p-3 sm:grid-cols-3 sm:gap-4 sm:p-4 lg:grid-cols-4 xl:grid-cols-5">
              {displayed.map((recipe) => (
                <RecipeCardDialog key={recipe.id} recipe={recipe} />
              ))}
            </div>
          )}
        </>
      )}
    </>
  )
}
