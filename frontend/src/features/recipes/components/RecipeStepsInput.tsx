import { joinSteps, splitSteps } from '@/features/recipes/steps'
import { Button } from '@/shared/ui/button'
import { Label } from '@/shared/ui/label'
import { Textarea } from '@/shared/ui/textarea'

type Props = {
  // procedure 文字列。単一の真実として親(RHF Controller)が保持する。
  value: string
  onChange: (value: string) => void
}

// 作り方を「手順1・手順2…」の追加方式で入力する。
// 値は procedure 文字列のまま親が持ち、本コンポーネントは手順配列との変換に徹する。
export function RecipeStepsInput({ value, onChange }: Props) {
  const parsed = splitSteps(value)
  // 最低 1 行は表示する(空でも手順1の入力欄を出す)。
  const steps = parsed.length > 0 ? parsed : ['']

  const update = (next: string[]) => onChange(joinSteps(next))

  const onChangeStep = (index: number, raw: string) => {
    // 1 行 = 1 手順で保持するため、手順内の改行は空白に潰す。
    const text = raw.replace(/\n/g, ' ')
    update(steps.map((s, i) => (i === index ? text : s)))
  }

  const onAddStep = () => update([...steps, ''])

  const onDropStep = (index: number) => {
    if (steps.length > 1) update(steps.filter((_, i) => i !== index))
  }

  return (
    <div className="grid gap-3">
      <Label>作り方</Label>
      <div className="flex flex-col gap-2">
        {steps.map((step, index) => (
          <div key={index} className="grid gap-1 rounded-md border p-2">
            <div className="flex items-center gap-2">
              <Label className="text-muted-foreground text-sm">手順{index + 1}</Label>
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="ml-auto"
                disabled={steps.length <= 1}
                onClick={() => onDropStep(index)}
              >
                -
              </Button>
            </div>
            <Textarea
              placeholder="手順を入力"
              value={step}
              onChange={(e) => onChangeStep(index, e.target.value)}
            />
          </div>
        ))}
      </div>
      <Button type="button" variant="outline" className="flex-1" onClick={onAddStep}>
        + 手順を追加
      </Button>
    </div>
  )
}
