import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm } from 'react-hook-form'

import {
  recipeFormSchema,
  type RecipeFormValues,
  toFormValues,
  toRecipeRequest,
} from '@/features/recipes/schema/recipeFormSchema'
import { INGREDIENT_UNITS, SEASONING_UNITS } from '@/features/recipes/units'
import type {
  LabelResponse,
  RecipeRequest,
  RecipeResponse,
  UserListItem,
} from '@/shared/api/generated/types.gen'
import { Button } from '@/shared/ui/button'
import { Checkbox } from '@/shared/ui/checkbox'
import { DialogClose, DialogFooter } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { Textarea } from '@/shared/ui/textarea'

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
  // フォーム状態は RHF + zod で一元管理する。検証は onBlur + onSubmit。
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
    <form onSubmit={submit} className="flex min-h-0 flex-1 flex-col gap-4">
      <div className="grid gap-4 overflow-auto pr-1">
        <div className="flex flex-col gap-3 sm:flex-row">
          <div className="grid flex-2 gap-3">
            <Label>タイトル</Label>
            <Controller
              control={control}
              name="title"
              render={({ field }) => (
                <Input
                  placeholder="タイトル"
                  value={field.value}
                  onChange={field.onChange}
                  onBlur={field.onBlur}
                />
              )}
            />
            {errors.title && <p className="text-destructive text-sm">{errors.title.message}</p>}
          </div>
          {/* スマホでは人数と調理時間を横並びにする。sm 以上では contents で解除し、
              タイトルと合わせて 3 つ横一列に戻す。 */}
          <div className="flex gap-3 sm:contents">
            <Controller
              control={control}
              name="createFor"
              render={({ field }) => (
                <SelectInput
                  className="grid flex-1 gap-3"
                  label="人数"
                  value={field.value}
                  onChange={field.onChange}
                  placeholder="選択"
                  options={[...Array(10)].map((_, i) => ({ label: `${i + 1}`, value: `${i + 1}` }))}
                />
              )}
            />
            <Controller
              control={control}
              name="createTime"
              render={({ field }) => (
                <SelectInput
                  className="grid flex-1 gap-3"
                  label="調理時間(分)"
                  value={field.value}
                  onChange={field.onChange}
                  placeholder="選択"
                  options={[...Array(30)].map((_, i) => {
                    const val = (i + 1) * 5
                    return { label: `${val}`, value: `${val}` }
                  })}
                />
              )}
            />
          </div>
        </div>
        <Controller
          control={control}
          name="ingredients"
          render={({ field }) => (
            <RecipeInputForm
              label="食材"
              value={field.value}
              onChange={field.onChange}
              units={INGREDIENT_UNITS}
              minRows={1}
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
              label="調味料"
              value={field.value}
              onChange={field.onChange}
              units={SEASONING_UNITS}
              minRows={0}
            />
          )}
        />
        <div className="grid gap-3">
          <Label>作り方</Label>
          <Controller
            control={control}
            name="procedure"
            render={({ field }) => (
              <Textarea placeholder="作り方を入力" value={field.value} onChange={field.onChange} />
            )}
          />
        </div>
        <div className="flex flex-col gap-3 sm:flex-row">
          {labelData && (
            <Controller
              control={control}
              name="label"
              render={({ field }) => (
                <MultiSelectInput
                  className="grid flex-1 gap-2"
                  label="ラベル"
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
                  className="grid flex-1 gap-2"
                  label="共有ユーザー"
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
          <Label htmlFor="archive_flg">アーカイブ</Label>
        </div>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button type="button" variant="outline" onClick={onClickCancel}>
            キャンセル
          </Button>
        </DialogClose>
        <Button type="submit">{mode === 'create' ? '作成' : '更新'}</Button>
      </DialogFooter>
    </form>
  )
}
