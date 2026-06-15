import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'

// 材料(食材・調味料)1 行の入力値。フォーム内部の状態型。
export type Material = { name: string; quantity: number; unit: string }

const emptyRow = (): Material => ({ name: '', quantity: 0, unit: '' })

type InputProps = {
  label: string
  value: Material[]
  onChange: (data: Material[]) => void
  // 残せる最小行数。食材は 1、調味料は 0(任意)。
  minRows?: number
}

// 値は親(RHF Controller)が単一の真実として保持し、本コンポーネントは制御コンポーネントに徹する。
// 内部 state + useEffect での逆流同期はしない。
export function RecipeInputForm({ label, value, onChange, minRows = 0 }: InputProps) {
  const onClickAddForm = () => {
    onChange([...value, emptyRow()])
  }

  const onClickDropForm = (index: number) => {
    if (value.length > minRows) {
      onChange(value.filter((_, i) => i !== index))
    }
  }

  const handleInputChange = (index: number, field: keyof Material, raw: string) => {
    const next = value.map((row, i) =>
      i === index ? { ...row, [field]: field === 'quantity' ? Number(raw) : raw } : row,
    )
    onChange(next)
  }

  return (
    <div className="grid gap-3">
      <Label>{label}</Label>
      <div className="flex flex-col gap-2">
        {value.map((material, index) => (
          <div key={index} className="flex gap-1">
            <div className="flex-2">
              <Input
                placeholder="name"
                value={material.name}
                onChange={(e) => handleInputChange(index, 'name', e.target.value)}
              />
            </div>
            <div className="flex-1">
              <Input
                type="number"
                placeholder="quantity"
                value={material.quantity}
                onChange={(e) => handleInputChange(index, 'quantity', e.target.value)}
              />
            </div>
            <div className="flex-1">
              <Input
                placeholder="unit"
                value={material.unit}
                onChange={(e) => handleInputChange(index, 'unit', e.target.value)}
              />
            </div>
            <div className="gap-1">
              <Button
                type="button"
                variant="outline"
                disabled={value.length <= minRows}
                onClick={() => onClickDropForm(index)}
              >
                -
              </Button>
            </div>
          </div>
        ))}
      </div>
      <div className="flex gap-2">
        <Button type="button" variant="outline" className="flex-1" onClick={onClickAddForm}>
          +
        </Button>
      </div>
    </div>
  )
}
