import { ChevronUp } from 'lucide-react'
import { type ReactNode, useState } from 'react'

import type { UnitConfig } from '@/features/recipes/units'
import { findUnit, formatAmount, quantityOptions } from '@/features/recipes/units'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/shared/ui/select'

// 材料(食材・調味料)1 行の入力値。フォーム内部の状態型。
export type Material = { name: string; quantity: number; unit: string }

const emptyRow = (): Material => ({ name: '', quantity: 0, unit: '' })

type InputProps = {
  label: string
  // 見出しに添えるアイコン(カード/詳細と統一)。
  icon?: ReactNode
  value: Material[]
  onChange: (data: Material[]) => void
  // この材料で選べる単位の一覧(食材/調味料で異なる)。
  units: UnitConfig[]
  // 残せる最小行数。食材は 1、調味料は 0(任意)。
  minRows?: number
}

// 材料はアコーディオン方式。入力済みの行は「名前 + 数量」のパネルに畳み、
// タップで開いて単位チップ・数量を編集する。開くのは常に 1 行(追加・他行を開くと畳む)。
export function RecipeInputForm({ label, icon, value, onChange, units, minRows = 0 }: InputProps) {
  // 未入力(単位なし)の行を最初だけ開く。全て入力済みなら畳んだ状態で始める。
  const [expanded, setExpanded] = useState<number>(() => value.findIndex((m) => m.unit === ''))
  // 開いている行で「その他(単位を手入力)」を選んでいるか。行を開き直すたびにリセットし、
  // 既存の手入力値(候補外の unit)は下の isCustomUnit で復元する。
  const [customOpen, setCustomOpen] = useState(false)

  // 行を開く。開くのは常に 1 行なので、その他モードも合わせて初期化する。
  const openRow = (index: number) => {
    setExpanded(index)
    setCustomOpen(false)
  }

  const updateRow = (index: number, patch: Partial<Material>) => {
    onChange(value.map((row, i) => (i === index ? { ...row, ...patch } : row)))
  }

  // 単位を選んだら、その単位の既定値に数量を合わせる(数量なし単位は 0)。
  const onSelectUnit = (index: number, config: UnitConfig) => {
    setCustomOpen(false)
    updateRow(index, { unit: config.unit, quantity: config.hasQuantity ? config.start : 0 })
  }

  // 「その他」を選ぶ。単位を手入力に切り替える。手入力は数値も単位テキストに含めるため、
  // 数量は持たない(適量・少々と同じ扱い)。既に手入力値がある行は消さない。
  const onSelectCustom = (index: number) => {
    setCustomOpen(true)
    if (units.some((u) => u.unit === value[index].unit)) {
      updateRow(index, { unit: '', quantity: 0 })
    }
  }

  const onAddForm = () => {
    onChange([...value, emptyRow()])
    openRow(value.length) // 追加した行を開く
  }

  const onDropForm = (index: number) => {
    if (value.length > minRows) {
      onChange(value.filter((_, i) => i !== index))
      openRow(-1)
    }
  }

  return (
    <div className="grid gap-3">
      <Label className="flex items-center gap-1">
        {icon}
        {label}
      </Label>
      <div className="flex flex-col gap-2">
        {value.map((material, index) => {
          const selected = findUnit(material.unit)
          const isOpen = expanded === index
          // その他(手入力)モード: この行で明示選択中か、候補外の値が既に入っている(編集時)。
          const isCustomUnit =
            (isOpen && customOpen) ||
            (material.unit !== '' && !units.some((u) => u.unit === material.unit))

          // 畳んだ状態: 名前 + 数量のパネル。タップで開く。
          if (!isOpen) {
            return (
              <div key={index} className="flex gap-1">
                <button
                  type="button"
                  onClick={() => openRow(index)}
                  className="flex flex-1 items-center gap-2 rounded-md border p-2 text-left"
                >
                  <span className="flex-1 truncate">{material.name || '（名前未入力）'}</span>
                  {material.unit && (
                    <span className="text-muted-foreground text-sm">
                      {formatAmount(material.quantity, material.unit)}
                    </span>
                  )}
                </button>
                <Button
                  type="button"
                  variant="outline"
                  disabled={value.length <= minRows}
                  onClick={() => onDropForm(index)}
                >
                  -
                </Button>
              </div>
            )
          }

          // 開いた状態: 名前 + 単位チップ + 数量。
          return (
            <div key={index} className="grid gap-2 rounded-md border p-2">
              <div className="flex gap-1">
                <Input
                  placeholder="名前"
                  value={material.name}
                  onChange={(e) => updateRow(index, { name: e.target.value })}
                />
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  aria-label="閉じる"
                  onClick={() => openRow(-1)}
                >
                  <ChevronUp />
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  disabled={value.length <= minRows}
                  onClick={() => onDropForm(index)}
                >
                  -
                </Button>
              </div>
              {/* 単位はタブ(チップ)で選ぶ。選択中は塗り、未選択は枠線。
                  末尾の「その他」で候補にない単位を手入力する。 */}
              <div className="flex flex-wrap gap-1">
                {units.map((config) => (
                  <Button
                    key={config.unit}
                    type="button"
                    size="sm"
                    variant={material.unit === config.unit ? 'default' : 'outline'}
                    onClick={() => onSelectUnit(index, config)}
                  >
                    {config.unit}
                  </Button>
                ))}
                <Button
                  type="button"
                  size="sm"
                  variant={isCustomUnit ? 'default' : 'outline'}
                  onClick={() => onSelectCustom(index)}
                >
                  その他
                </Button>
              </div>
              {/* その他は単位を手入力(数値も単位に含める)。候補の単位は数量を候補から選ぶ。
                  数量なし単位(適量・少々)では数量欄を出さない。 */}
              {isCustomUnit ? (
                <Input
                  placeholder="単位を入力（例: 3房、ひとつまみ）"
                  maxLength={10}
                  value={material.unit}
                  onChange={(e) => updateRow(index, { unit: e.target.value })}
                />
              ) : (
                selected?.hasQuantity && (
                  <Select
                    value={String(material.quantity)}
                    onValueChange={(v) => updateRow(index, { quantity: Number(v) })}
                  >
                    <SelectTrigger className="w-32">
                      <SelectValue placeholder="数量" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {quantityOptions(selected).map((opt) => (
                          <SelectItem key={opt.value} value={opt.value}>
                            {opt.label}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                )
              )}
            </div>
          )
        })}
      </div>
      <Button type="button" variant="outline" className="flex-1" onClick={onAddForm}>
        + {label}を追加
      </Button>
    </div>
  )
}
