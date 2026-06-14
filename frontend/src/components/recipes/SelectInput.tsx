import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectGroup,
  SelectItem,
} from "@/components/ui/select"
import { Label } from "@/components/ui/label"

type SelectInputProps = {
  className: string;
  label: string
  value: string
  placeholder?: string
  onChange: (value: string) => void
  options: Array<{ label: string; value: string }>
}

export function SelectInput({
  className,
  label,
  value,
  placeholder = "選択してください",
  onChange,
  options,
}: SelectInputProps) {
  return (
    <div className={className}>
      <Label>{label}</Label>
      <Select value={value} onValueChange={onChange}>
        <SelectTrigger className="w-full">
          <SelectValue placeholder={placeholder} />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            {options.map((option, i) => (
              <SelectItem key={i} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectGroup>
        </SelectContent>
      </Select>
    </div>
  )
}