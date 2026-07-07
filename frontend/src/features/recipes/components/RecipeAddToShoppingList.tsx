import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ShoppingCart } from 'lucide-react'
import { useMemo, useState } from 'react'
import { toast } from 'sonner'

import { formatAmount } from '@/features/recipes/units'
import {
  addShoppingListItemsMutation,
  getShoppingListOptions,
  getShoppingListQueryKey,
} from '@/shared/api/generated/@tanstack/react-query.gen'
import type { RecipeResponse, ShoppingListBulkAddItem } from '@/shared/api/generated/types.gen'
import { Button } from '@/shared/ui/button'
import { Checkbox } from '@/shared/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/shared/ui/dialog'

// 選択候補の 1 行。材料と調味料を同じ形に揃え、key で種別+名前を一意にする。
type Selectable = { key: string; name: string; quantity: number; unit: string }

// レシピの材料・調味料から、選んだものを買い物リストへ一括追加するボタン + 選択ダイアログ。
// 数量・単位はそのまま持ち込み、重複はサーバ側でマージせず別行で追加される。
export function RecipeAddToShoppingList({ recipe }: { recipe: RecipeResponse }) {
  const [open, setOpen] = useState(false)
  const [selected, setSelected] = useState<Set<string>>(() => new Set())
  const queryClient = useQueryClient()

  // 追加先は取得ユーザーの買い物リスト(共有時はグループの共有リスト)。キャッシュ済みを使う。
  const { data: list } = useQuery(getShoppingListOptions())

  const candidates = useMemo<Selectable[]>(() => {
    // key は種別 + 並び順で一意にする(同名の材料/調味料でも別行として扱える)。
    const cooking = recipe.cooking.map((c, i) => ({
      key: `c:${i}`,
      name: c.ingredients.name,
      quantity: c.quantity,
      unit: c.unit,
    }))
    const season = recipe.season.map((s, i) => ({
      key: `s:${i}`,
      name: s.seasoning.name,
      quantity: s.quantity,
      unit: s.unit,
    }))
    return [...cooking, ...season]
  }, [recipe])

  const addItems = useMutation({
    ...addShoppingListItemsMutation(),
    onSuccess: (data) => {
      // 追加後のリスト全体が返るのでキャッシュを差し替える(買い物リスト画面に即反映)。
      queryClient.setQueryData(getShoppingListQueryKey(), data)
      toast.success('買い物リストに追加しました')
      setOpen(false)
    },
    onError: () => toast.error('買い物リストへの追加に失敗しました'),
  })

  // 開くたびに全選択で初期化する。
  const handleOpenChange = (next: boolean) => {
    if (next) setSelected(new Set(candidates.map((c) => c.key)))
    setOpen(next)
  }

  const toggle = (key: string, checked: boolean) =>
    setSelected((prev) => {
      const next = new Set(prev)
      if (checked) next.add(key)
      else next.delete(key)
      return next
    })

  const handleAdd = () => {
    if (!list) return
    const items: ShoppingListBulkAddItem[] = candidates
      .filter((c) => selected.has(c.key))
      .map((c) => ({ name: c.name, quantity: c.quantity, unit: c.unit }))
    if (items.length === 0) return
    addItems.mutate({ path: { id: list.id }, body: { items } })
  }

  // 材料も調味料も無いレシピでは出さない。
  if (candidates.length === 0) return null

  // selected は候補の key しか持たないので件数は size と一致する。
  const selectedCount = selected.size

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button variant="outline" className="w-full" disabled={!list}>
          <ShoppingCart />
          買い物リストに追加
        </Button>
      </DialogTrigger>
      <DialogContent className="flex max-h-[85dvh] flex-col">
        <DialogHeader>
          <DialogTitle>買い物リストに追加</DialogTitle>
          <DialogDescription>追加する材料・調味料を選んでください。</DialogDescription>
        </DialogHeader>
        <div className="flex-1 overflow-auto">
          <ul className="divide-border divide-y">
            {candidates.map((c) => (
              <li key={c.key}>
                <label className="flex cursor-pointer items-center gap-3 py-2.5">
                  <Checkbox
                    checked={selected.has(c.key)}
                    onCheckedChange={(checked) => toggle(c.key, checked === true)}
                  />
                  <span className="min-w-0 flex-1 truncate text-sm">{c.name}</span>
                  <span className="text-muted-foreground shrink-0 text-sm tabular-nums">
                    {formatAmount(c.quantity, c.unit)}
                  </span>
                </label>
              </li>
            ))}
          </ul>
        </div>
        <DialogFooter>
          <Button onClick={handleAdd} disabled={addItems.isPending || selectedCount === 0 || !list}>
            {selectedCount}件を追加
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
