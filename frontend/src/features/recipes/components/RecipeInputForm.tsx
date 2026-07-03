import type { UnitConfig } from '@/features/recipes/units'
import { findUnit, quantityOptions } from '@/features/recipes/units'
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
  value: Material[]
  onChange: (data: Material[]) => void
  // この材料で選べる単位の一覧(食材/調味料で異なる)。
  units: UnitConfig[]
  // 残せる最小行数。食材は 1、調味料は 0(任意)。
  minRows?: number
}

// 値は親(RHF Controller)が単一の真実として保持し、本コンポーネントは制御コンポーネントに徹する。
export function RecipeInputForm({ label, value, onChange, units, minRows = 0 }: InputProps) {
  const onClickAddForm = () => {
    onChange([...value, emptyRow()])
  }

  const onClickDropForm = (index: number) => {
    if (value.length > minRows) {
      onChange(value.filter((_, i) => i !== index))
    }
  }

  const updateRow = (index: number, patch: Partial<Material>) => {
    onChange(value.map((row, i) => (i === index ? { ...row, ...patch } : row)))
  }

  // 単位を選んだら、その単位の既定値に数量を合わせる(数量なし単位は 0)。
  const onSelectUnit = (index: number, config: UnitConfig) => {
    updateRow(index, { unit: config.unit, quantity: config.hasQuantity ? config.start : 0 })
  }

  return (
    <div className="grid gap-3">
      <Label>{label}</Label>
      <div className="flex flex-col gap-3">
        {value.map((material, index) => {
          const selected = findUnit(material.unit)
          const showQuantity = selected?.hasQuantity ?? false
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
                  disabled={value.length <= minRows}
                  onClick={() => onClickDropForm(index)}
                >
                  -
                </Button>
              </div>
              {/* 単位はタブ(チップ)で選ぶ。選択中は塗り、未選択は枠線。 */}
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
              </div>
              {/* 数量は単位に応じた候補から選ぶ。数量なし単位(適量・少々)では出さない。 */}
              {showQuantity && selected && (
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
              )}
            </div>
          )
        })}
      </div>
      <div className="flex gap-2">
        <Button type="button" variant="outline" className="flex-1" onClick={onClickAddForm}>
          +
        </Button>
      </div>
    </div>
  )
}
