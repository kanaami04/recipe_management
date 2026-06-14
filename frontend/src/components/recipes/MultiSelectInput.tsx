import { Label } from "@/components/ui/label"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Popover, PopoverTrigger, PopoverContent } from "@/components/ui/popover"

type MultiSelectInputProps = {
  className?: string;
  label?: string
  value: string[]
  placeholder?: string
  onChange: (value: string[]) => void
  options: { label: string; value: string }[]
}

export function MultiSelectInput({
  className = "",
  label,
  value,
  placeholder = "選択してください",
  onChange,
  options,
}: MultiSelectInputProps) {
  const toggle = (v: string) => {
    if (value.includes(v)) onChange(value.filter((x) => x !== v))
    else onChange([...value, v])
  }

  return (
    <div className={className}>
      {label && <Label>{label}</Label>}
      <Popover>
        <PopoverTrigger asChild>
          <Button variant="outline" className="w-full text-left">
            <div>
              {value.length === 0 ? (
                <span className="text-muted-foreground">{placeholder}</span>
              ) : (
                value.map((item) => {
                  const opt = options.find((o) => o.value === item)
                  return (
                    <span key={item} className="inline-flex items-center gap-1 px-2 py-1 rounded-full border text-sm">
                      {opt?.label ?? value}
                    </span>
                  )
                })
              )}
            </div>
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-64 p-2">
          <div className="flex flex-col gap-2 max-h-60 overflow-auto">
            {options.map((opt) => (
              <label
                key={opt.value}
                className="flex items-center gap-2 cursor-pointer px-1 py-1 rounded hover:bg-muted"
              >
                <Checkbox checked={value.includes(opt.value)} onCheckedChange={() => toggle(opt.value)} />
                <span>{opt.label}</span>
              </label>
            ))}
          </div>
        </PopoverContent>
      </Popover>
    </div>
  )
}