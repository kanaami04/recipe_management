import { zodResolver } from '@hookform/resolvers/zod'
import { Carrot, Clock, Droplet, Users } from 'lucide-react'
import { Controller, useForm } from 'react-hook-form'

import {
  recipeFormSchema,
  type RecipeFormValues,
  type RecipeInputMode,
  toFormValues,
  toRecipeRequest,
} from '@/features/recipes/schema/recipeFormSchema'
import { INGREDIENT_UNITS, SEASONING_UNITS } from '@/features/recipes/units'
import type {
  LabelItem,
  RecipeRequest,
  RecipeResponse,
  UserListItem,
} from '@/shared/api/generated/types.gen'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/ui/button'
import { DialogFooter } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { Separator } from '@/shared/ui/separator'

import { MultiSelectInput } from './MultiSelectInput'
import { RecipeInputForm } from './RecipeInputForm'
import { RecipeSourceInput } from './RecipeSourceInput'
import { RecipeStepsInput } from './RecipeStepsInput'
import { SelectInput } from './SelectInput'

type Props = {
  mode: 'create' | 'edit'
  initialData?: RecipeResponse
  onSubmit: (payload: RecipeRequest) => Promise<void>
  labelData?: LabelItem[]
  sharedUserData?: UserListItem[]
}

export function RecipeForm({ mode, initialData, onSubmit, labelData, sharedUserData }: Props) {
  // フォーム状態は RHF + zod で一元管理する。
  // 検証は「作成/更新」ボタン押下時(onSubmit)に行い、必須項目の警告もそこで出す。
  // 一度送信した後は onChange で再検証され、入力を直すと警告が即座に消える(RHF 既定)。
  const {
    control,
    handleSubmit,
    setValue,
    watch,
    clearErrors,
    formState: { errors },
  } = useForm<RecipeFormValues>({
    resolver: zodResolver(recipeFormSchema),
    defaultValues: toFormValues(initialData),
    mode: 'onSubmit',
  })

  // サムネイルは URL 入力から OGP 取得して programmatic に埋めるため watch で参照する。
  const thumbnailUrl = watch('thumbnailUrl')
  // 入力モード。url モードでは URL とサムネだけで登録でき、手動入力の必須項目は問わない。
  const inputMode = watch('inputMode')

  const changeMode = (next: RecipeInputMode) => {
    setValue('inputMode', next)
    clearErrors() // モードで必須条件が変わるため、残っている警告は消す。
  }

  const submit = handleSubmit(async (values) => {
    await onSubmit(toRecipeRequest(values))
  })

  return (
    <form onSubmit={submit} className="flex min-h-0 flex-1 flex-col gap-4">
      {/* overflow-auto は子の box-shadow を切り落とすため、左右対称の余白(px-1)で
          タイトル等の Input の影が片側だけ欠けないようにする。 */}
      <div className="scrollbar-none grid gap-4 overflow-auto px-1">
        <ModeToggle value={inputMode} onChange={changeMode} />

        {inputMode === 'url' ? (
          <>
            {/* url モードでは参考 URL とサムネ、自動取得のレシピ名だけで登録できる。 */}
            <Controller
              control={control}
              name="sourceUrl"
              render={({ field }) => (
                <RecipeSourceInput
                  url={field.value}
                  thumbnail={thumbnailUrl}
                  onUrlChange={field.onChange}
                  onThumbnailChange={(value) =>
                    setValue('thumbnailUrl', value, { shouldDirty: true })
                  }
                  // レシピ名を OGP タイトルで自動補完する。
                  onTitleFetched={(title) => setValue('title', title, { shouldDirty: true })}
                />
              )}
            />
            {errors.sourceUrl && (
              <p className="text-destructive text-sm">{errors.sourceUrl.message}</p>
            )}
            <div className="grid gap-3">
              <Label>レシピ名(自動取得・編集できます)</Label>
              <Controller
                control={control}
                name="title"
                render={({ field }) => (
                  <Input
                    placeholder="URL から自動で取得します"
                    value={field.value}
                    onChange={field.onChange}
                    onBlur={field.onBlur}
                  />
                )}
              />
            </div>
          </>
        ) : (
          <>
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
                      label={
                        <>
                          <Users className="size-4" />
                          人数
                        </>
                      }
                      value={field.value}
                      onChange={field.onChange}
                      placeholder="選択"
                      options={[...Array(10)].map((_, i) => ({
                        label: `${i + 1}`,
                        value: `${i + 1}`,
                      }))}
                    />
                  )}
                />
                <Controller
                  control={control}
                  name="createTime"
                  render={({ field }) => (
                    <SelectInput
                      className="grid flex-1 gap-3"
                      label={
                        <>
                          <Clock className="size-4" />
                          調理時間(分)
                        </>
                      }
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
            {errors.createFor && (
              <p className="text-destructive text-sm">{errors.createFor.message}</p>
            )}
            <Separator />
            <Controller
              control={control}
              name="ingredients"
              render={({ field }) => (
                <RecipeInputForm
                  label="食材"
                  icon={<Carrot className="size-4" />}
                  value={field.value}
                  onChange={field.onChange}
                  units={INGREDIENT_UNITS}
                  minRows={1}
                />
              )}
            />
            {errors.ingredients && (
              <p className="text-destructive text-sm">
                {errors.ingredients.message ?? '食材の名前と単位を入力してください'}
              </p>
            )}
            <Separator />
            <Controller
              control={control}
              name="seasoning"
              render={({ field }) => (
                <RecipeInputForm
                  label="調味料"
                  icon={<Droplet className="size-4" />}
                  value={field.value}
                  onChange={field.onChange}
                  units={SEASONING_UNITS}
                  minRows={0}
                />
              )}
            />
            {errors.seasoning && (
              <p className="text-destructive text-sm">
                {errors.seasoning.message ?? '調味料の名前と単位を入力してください'}
              </p>
            )}
            <Separator />
            <Controller
              control={control}
              name="procedure"
              render={({ field }) => (
                <RecipeStepsInput value={field.value} onChange={field.onChange} />
              )}
            />
          </>
        )}
        <Separator />
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
      </div>
      <DialogFooter>
        {/* キャンセルは DialogContent 既定の×と重複するため置かない。 */}
        <Button type="submit">{mode === 'create' ? '作成' : '更新'}</Button>
      </DialogFooter>
    </form>
  )
}

// 入力モードの切り替え(手動入力 / URL から登録)。segmented control 風の 2 択トグル。
function ModeToggle({
  value,
  onChange,
}: {
  value: RecipeInputMode
  onChange: (mode: RecipeInputMode) => void
}) {
  const options: { key: RecipeInputMode; label: string }[] = [
    { key: 'manual', label: '手動で入力' },
    { key: 'url', label: 'URL から登録' },
  ]
  return (
    <div className="bg-muted grid grid-cols-2 gap-1 rounded-lg p-1">
      {options.map((option) => (
        <button
          key={option.key}
          type="button"
          onClick={() => onChange(option.key)}
          className={cn(
            'rounded-md py-1.5 text-sm transition',
            value === option.key
              ? 'bg-background text-foreground font-medium shadow-sm'
              : 'text-muted-foreground',
          )}
        >
          {option.label}
        </button>
      ))}
    </div>
  )
}
