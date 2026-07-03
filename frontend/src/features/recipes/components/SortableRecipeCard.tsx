import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { GripVertical } from 'lucide-react'

import { RecipeCardDialog } from '@/features/recipes/components/RecipeCardDialog'
import type { RecipeResponse } from '@/shared/api/generated/types.gen'

// 並び替え可能なレシピカード。カード全体がドラッグの掴み代。
// タップ(詳細を開く)・スクロールとの競合は、マウス=距離しきい値/タッチ=長押しの
// センサー設定(RecipesPage)側で解決する。右上のグリップは掴めることを示す装飾。
export function SortableRecipeCard({ recipe }: { recipe: RecipeResponse }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: recipe.id,
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.4 : undefined,
    zIndex: isDragging ? 10 : undefined,
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      className="relative cursor-grab active:cursor-grabbing"
      {...attributes}
      {...listeners}
    >
      <GripVertical className="text-muted-foreground pointer-events-none absolute top-1 right-1 z-10 size-4 opacity-50" />
      <RecipeCardDialog recipe={recipe} />
    </div>
  )
}
