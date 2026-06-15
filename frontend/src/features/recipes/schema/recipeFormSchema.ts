import { z } from 'zod'

import type { RecipeRequest, RecipeResponse } from '@/shared/api/generated/types.gen'

// フォーム用の zod スキーマ(手書き)。UI に素直な形にする (ADR-0006)。
// API レスポンス用 zod は生成物を使い、住み分ける (ADR-0007)。
const materialSchema = z.object({
  name: z.string().min(1, '名前は必須です'),
  quantity: z.number().int().min(0),
  unit: z.string(),
})

export const recipeFormSchema = z.object({
  title: z.string().min(1, 'タイトルは必須です'),
  createFor: z.string().min(1, '人数を選択してください'),
  createTime: z.string(), // 任意。空文字を許容し、変換時に null へ。
  procedure: z.string(),
  archiveFlg: z.boolean(),
  label: z.array(z.string()),
  sharedUser: z.array(z.string()),
  ingredients: z.array(materialSchema).min(1, '食材は1つ以上必要です'),
  seasoning: z.array(materialSchema),
})

export type RecipeFormValues = z.infer<typeof recipeFormSchema>

// API レスポンス(RecipeResponse)→ フォーム値。defaultValues に使う。
export function toFormValues(recipe?: RecipeResponse): RecipeFormValues {
  return {
    title: recipe?.title ?? '',
    createFor: recipe ? String(recipe.create_for) : '',
    createTime: recipe?.create_time != null ? String(recipe.create_time) : '',
    procedure: recipe?.procedure ?? '',
    archiveFlg: recipe?.archive_flg ?? false,
    label: recipe?.label.map((l) => l.name) ?? [],
    sharedUser: recipe?.shared_user.map((u) => u.username) ?? [],
    ingredients: recipe?.cooking.map((c) => ({
      name: c.ingredients.name,
      quantity: c.quantity,
      unit: c.unit,
    })) ?? [{ name: '', quantity: 0, unit: '' }],
    seasoning: recipe?.season.map((s) => ({
      name: s.seasoning.name,
      quantity: s.quantity,
      unit: s.unit,
    })) ?? [{ name: '', quantity: 0, unit: '' }],
  }
}

// フォーム値 → API リクエスト(RecipeRequest)。id は URL 側で扱う。
export function toRecipeRequest(values: RecipeFormValues): RecipeRequest {
  return {
    title: values.title,
    create_for: Number(values.createFor),
    create_time: values.createTime === '' ? null : Number(values.createTime),
    procedure: values.procedure,
    archive_flg: values.archiveFlg,
    label: values.label.map((name) => ({ name })),
    shared_user: values.sharedUser.map((username) => ({ username })),
    cooking: values.ingredients.map((m) => ({
      ingredients: { name: m.name },
      quantity: m.quantity,
      unit: m.unit,
    })),
    season: values.seasoning.map((m) => ({
      seasoning: { name: m.name },
      quantity: m.quantity,
      unit: m.unit,
    })),
  }
}
