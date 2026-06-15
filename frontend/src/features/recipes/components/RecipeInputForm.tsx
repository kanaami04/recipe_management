import { useEffect, useState } from 'react'

import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'

// 材料(食材・調味料)1 行の入力値。フォーム内部の状態型。
export type Material = { name: string; quantity: number; unit: string }

type InputProps = {
  label: string
  initialInputData?: Material[] | null
  onChange: (data: Material[]) => void
}

export function RecipeInputForm({ label, initialInputData, onChange }: InputProps) {
  const [inputs, setInputs] = useState<Material[]>(
    initialInputData ?? [{ name: '', quantity: 0, unit: '' }],
  )

  useEffect(() => {
    onChange(inputs)
  }, [inputs, onChange])

  const onClickAddForm = () => {
    setInputs((prevInputs) => [...prevInputs, { name: '', quantity: 0, unit: '' }])
  }

  const onClickDropForm = (index: number) => {
    if (inputs.length > 1) {
      setInputs((prevInputs) => prevInputs.filter((_, i) => i !== index))
    }
  }

  const handleInputChange = (index: number, field: keyof Material, value: string) => {
    const newInputs = [...inputs]
    newInputs[index] = {
      ...newInputs[index],
      [field]: field === 'quantity' ? Number(value) : value,
    }
    setInputs(newInputs)
  }

  return (
    <div className="grid gap-3">
      <Label>{label}</Label>
      <div className="flex flex-col gap-2">
        {inputs.map((material, index) => (
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
              <Button className="flex-1" onClick={() => onClickDropForm(index)}>
                -
              </Button>
            </div>
          </div>
        ))}
      </div>
      <div className="flex gap-2">
        <Button className="flex-1" onClick={onClickAddForm}>
          +
        </Button>
      </div>
    </div>
  )
}
