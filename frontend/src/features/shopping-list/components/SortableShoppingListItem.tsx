import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { GripVertical, Trash2 } from 'lucide-react'

import type { ShoppingListItemResponse } from '@/shared/api/generated/types.gen'
import { Button } from '@/shared/ui/button'
import { Checkbox } from '@/shared/ui/checkbox'

type SortableShoppingListItemProps = {
  item: ShoppingListItemResponse
  onToggle: (checked: boolean) => void
  onDelete: () => void
}

// 並び替え可能な未チェック項目。左端のグリップだけがドラッグの掴み代で、チェックボックス・
// 削除ボタンのタップとは干渉しない。長押し(タッチ)・距離しきい値(マウス)での開始は
// ShoppingListPage 側のセンサー設定で解決する。
export function SortableShoppingListItem({
  item,
  onToggle,
  onDelete,
}: SortableShoppingListItemProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: item.id,
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.4 : undefined,
    zIndex: isDragging ? 10 : undefined,
  }

  return (
    <li ref={setNodeRef} style={style} className="bg-card flex items-center gap-2 p-3">
      {/* 掴み代。touch-none でタッチ中のスクロールと衝突させない。 */}
      <button
        type="button"
        aria-label="並び替え"
        className="text-muted-foreground touch-none cursor-grab active:cursor-grabbing"
        {...attributes}
        {...listeners}
      >
        <GripVertical className="size-5" />
      </button>
      <label className="flex flex-1 cursor-pointer items-center gap-3">
        <Checkbox
          checked={item.checked}
          onCheckedChange={(checked) => onToggle(checked === true)}
        />
        <span className="truncate">{item.name}</span>
      </label>
      <Button variant="ghost" size="icon" aria-label="削除" onClick={onDelete}>
        <Trash2 />
      </Button>
    </li>
  )
}
