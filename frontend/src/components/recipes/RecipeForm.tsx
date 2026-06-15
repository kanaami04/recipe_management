import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm } from 'react-hook-form'

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

import { MultiSelectInput } from './MultiSelectInput'
import {
  recipeFormSchema,
  type RecipeFormValues,
  toFormValues,
  toRecipeRequest,
} from './recipeFormSchema'
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
  // フォーム状態は RHF + zod で一元管理する (ADR-0006)。検証は onBlur + onSubmit。
  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<RecipeFormValues>({
    resolver: zodResolver(recipeFormSchema),
    defaultValues: toFormValues(initialData),
    mode: 'onBlur',
  })

  const submit = handleSubmit(async (values) => {
    await onSubmit(toRecipeRequest(values))
  })

  return (
    <form onSubmit={submit}>
      <div className="grid gap-4 h-140 overflow-auto">
        <div className="flex gap-3">
          <div className="flex-2 grid gap-3">
            <Label>title</Label>
            <Controller
              control={control}
              name="title"
              render={({ field }) => (
                <Input
                  placeholder="title"
                  value={field.value}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                />
              )}
            />
            {errors.title && <p className="text-destructive text-sm">{errors.title.message}</p>}
          </div>
          <Controller
            control={control}
            name="createFor"
            render={({ field }) => (
              <SelectInput
                className="flex-1 grid gap-3"
                label="create_for"
                value={field.value}
                onChange={field.onChange}
                placeholder="number"
                options={[...Array(10)].map((_, i) => ({ label: `${i + 1}`, value: `${i + 1}` }))}
              />
            )}
          />
          <Controller
            control={control}
            name="createTime"
            render={({ field }) => (
              <SelectInput
                className="flex-1 w-20 grid gap-3"
                label="create_time"
                value={field.value}
                onChange={field.onChange}
                placeholder="time"
                options={[...Array(30)].map((_, i) => {
                  const val = (i + 1) * 5
                  return { label: `${val}`, value: `${val}` }
                })}
              />
            )}
          />
        </div>
        <Controller
          control={control}
          name="ingredients"
          render={({ field }) => (
            <RecipeInputForm
              label="ingredients"
              initialInputData={mode === 'create' ? null : field.value}
              onChange={field.onChange}
            />
          )}
        />
        {errors.ingredients && (
          <p className="text-destructive text-sm">{errors.ingredients.message}</p>
        )}
        <Controller
          control={control}
          name="seasoning"
          render={({ field }) => (
            <RecipeInputForm
              label="seasoning"
              initialInputData={mode === 'create' ? null : field.value}
              onChange={field.onChange}
            />
          )}
        />
        <div className="grid gap-3">
          <Label>procedure</Label>
          <Controller
            control={control}
            name="procedure"
            render={({ field }) => (
              <Textarea placeholder="..." value={field.value} onChange={field.onChange} />
            )}
          />
        </div>
        <div className="flex gap-3">
          {labelData && (
            <Controller
              control={control}
              name="label"
              render={({ field }) => (
                <MultiSelectInput
                  className="flex-1 w-20 grid gap-2"
                  label="label"
                  value={field.value}
                  onChange={field.onChange}
                  options={labelData.map((l) => ({ label: l.name, value: l.name }))}
                />
              )}
            />
          )}
          {sharedUserData && (
            <Controller
              control={control}
              name="sharedUser"
              render={({ field }) => (
                <MultiSelectInput
                  className="flex-1 w-20 grid gap-2"
                  label="shared"
                  value={field.value}
                  onChange={field.onChange}
                  options={sharedUserData.map((u) => ({ label: u.username, value: u.username }))}
                />
              )}
            />
          )}
        </div>
        <div className="flex items-center gap-3">
          <Controller
            control={control}
            name="archiveFlg"
            render={({ field }) => (
              <Checkbox
                id="archive_flg"
                checked={field.value}
                onCheckedChange={(value) => field.onChange(Boolean(value))}
              />
            )}
          />
          <Label htmlFor="archive_flg">archive</Label>
        </div>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button type="button" variant="outline" onClick={onClickCancel}>
            Cancel
          </Button>
        </DialogClose>
        <Button type="submit">{mode === 'create' ? 'Create' : 'Update'}</Button>
      </DialogFooter>
    </form>
  )
}
