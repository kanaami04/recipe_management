import { useState } from 'react'

import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { DialogClose, DialogFooter } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import type {
  LabelResponse,
  RecipeRequest,
  RecipeResponse,
  UserListItem,
} from '@/shared/api/generated/types.gen'
import type { Material } from '@/type/RecipeDataType'

import { MultiSelectInput } from './MultiSelectInput'
import { RecipeInputForm } from './RecipeInputForm'
import { SelectInput } from './SelectInput'

type Props = {
  mode: 'create' | 'edit'
  initialData?: RecipeResponse
  onSubmit: (payload: RecipeRequest) => Promise<void>
  labelData?: LabelResponse[]
  sharedUserData?: UserListItem[]
  onClickCancel?: () => void
}

export function RecipeForm({
  mode,
  initialData,
  onSubmit,
  labelData,
  sharedUserData,
  onClickCancel,
}: Props) {
  const [title, setTitle] = useState(initialData?.title ?? '')
  const [createFor, setCreateFor] = useState<string>(String(initialData?.create_for ?? ''))
  const [time, setTime] = useState<string>(String(initialData?.create_time ?? ''))
  const [ingredients, setIngredients] = useState<Material[]>(
    initialData?.cooking
      ? initialData?.cooking.map((value) => ({
          name: value.ingredients.name,
          quantity: value.quantity,
          unit: value.unit,
        }))
      : [],
  )
  const [seasoning, setSeasoning] = useState<Material[]>(
    initialData?.season
      ? initialData?.season.map((value) => ({
          name: value.seasoning.name,
          quantity: value.quantity,
          unit: value.unit,
        }))
      : [],
  )
  const [procedure, setProcedure] = useState<string>(initialData?.procedure ?? '')
  const [isArchive, setISArchive] = useState(initialData?.archive_flg ?? false)
  const [label, setLabel] = useState<string[]>(initialData?.label?.map((l) => l.name) ?? [])
  const [sharedUser, setSharedUser] = useState<string[]>(
    initialData?.shared_user?.map((s) => s.username) ?? [],
  )

  const handleCreatePayload: React.FormEventHandler<HTMLFormElement> = async (e) => {
    e.preventDefault()

    // フォームの状態を API 契約(RecipeRequest)へ変換する。id は URL 側で扱う。
    const payload: RecipeRequest = {
      title,
      create_time: time === '' ? null : Number(time),
      create_for: Number(createFor),
      procedure,
      archive_flg: isArchive,
      label: label.map((l) => ({ name: l })),
      shared_user: sharedUser.map((s) => ({ username: s })),
      cooking: ingredients.map((item) => ({
        ingredients: { name: item.name },
        quantity: item.quantity,
        unit: item.unit,
      })),
      season: seasoning.map((item) => ({
        seasoning: { name: item.name },
        quantity: item.quantity,
        unit: item.unit,
      })),
    }

    await onSubmit(payload)
  }

  return (
    <form onSubmit={handleCreatePayload}>
      <div className="grid gap-4 h-140 overflow-auto">
        <div className="flex gap-3">
          <div className="flex-2 grid gap-3">
            <Label>title</Label>
            <Input placeholder="title" value={title} onChange={(e) => setTitle(e.target.value)} />
          </div>
          <SelectInput
            className="flex-1 grid gap-3"
            label="create_for"
            value={createFor}
            onChange={setCreateFor}
            placeholder="number"
            options={[...Array(10)].map((_, i) => ({
              label: `${i + 1}`,
              value: `${i + 1}`,
            }))}
          />
          <SelectInput
            className="flex-1 w-20 grid gap-3"
            label="create_time"
            value={time}
            onChange={setTime}
            placeholder="time"
            options={[...Array(30)].map((_, i) => {
              const val = (i + 1) * 5
              return { label: `${val}`, value: `${val}` }
            })}
          />
        </div>
        <RecipeInputForm
          label="ingredients"
          initialInputData={mode === 'create' ? null : ingredients}
          onChange={setIngredients}
        />
        <RecipeInputForm
          label="seasoning"
          initialInputData={mode === 'create' ? null : seasoning}
          onChange={setSeasoning}
        />
        <div className="grid gap-3">
          <Label>procedure</Label>
          <Textarea
            placeholder="..."
            value={procedure}
            onChange={(e) => setProcedure(e.target.value)}
          />
        </div>
        <div className="flex gap-3">
          {labelData && (
            <MultiSelectInput
              className="flex-1 w-20 grid gap-2"
              label="label"
              value={label}
              onChange={setLabel}
              options={labelData.map((label) => ({
                label: label.name,
                value: label.name,
              }))}
            />
          )}
          {sharedUserData && (
            <MultiSelectInput
              className="flex-1 w-20 grid gap-2"
              label="shared"
              value={sharedUser}
              onChange={setSharedUser}
              options={sharedUserData.map((user) => ({
                label: user.username,
                value: user.username,
              }))}
            />
          )}
        </div>
        <div className="flex items-center gap-3">
          <Checkbox
            id="archive_flg"
            checked={isArchive}
            onCheckedChange={(value: boolean) => setISArchive(value)}
          />
          <Label htmlFor="archive_flg">archive</Label>
        </div>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="outline" onClick={onClickCancel}>
            Cancel
          </Button>
        </DialogClose>
        <Button type="submit">{mode === 'create' ? 'Create' : 'Update'}</Button>
      </DialogFooter>
    </form>
  )
}
