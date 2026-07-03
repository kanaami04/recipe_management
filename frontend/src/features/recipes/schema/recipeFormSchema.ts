import { z } from 'zod'

import type { RecipeRequest, RecipeResponse } from '@/shared/api/generated/types.gen'

// フォーム用の zod スキーマ(手書き)。UI に素直な形にする。
// API レスポンス用 zod は生成物を使い、住み分ける。
const materialSchema = z.object({
  name: z.string(),
  quantity: z.number().min(0),
  unit: z.string(),
})

type Material = z.infer<typeof materialSchema>

// 名前・単位ともに未入力の行は「未使用」とみなす。初期表示される空行や、
// 追加したまま埋めなかった行は、検証・送信の対象から外す。
const isEmptyRow = (m: Material): boolean => m.name === '' && m.unit === ''
const isCompleteRow = (m: Material): boolean => m.name !== '' && m.unit !== ''

// 未使用でない行(名前か単位を入力した行)は、名前・単位ともに必須。
const validateRows = (rows: Material[], ctx: z.RefinementCtx) => {
  rows.forEach((row, index) => {
    if (isEmptyRow(row)) return
    if (row.name === '') {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: [index, 'name'],
        message: '名前は必須です',
      })
    }
    if (row.unit === '') {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: [index, 'unit'],
        message: '単位を選択してください',
      })
    }
  })
}

export const recipeFormSchema = z.object({
  title: z.string().min(1, 'タイトルは必須です'),
  createFor: z.string().min(1, '人数を選択してください'),
  createTime: z.string(), // 任意。空文字を許容し、変換時に null へ。
  procedure: z.string(),
  archiveFlg: z.boolean(),
  label: z.array(z.string()),
  sharedUser: z.array(z.string()),
  // 食材は完全に入力された行が最低 1 つ必須。調味料は任意(空行は無視)。
  ingredients: z
    .array(materialSchema)
    .superRefine(validateRows)
    .refine((rows) => rows.some(isCompleteRow), { message: '食材は1つ以上必要です' }),
  seasoning: z.array(materialSchema).superRefine(validateRows),
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
    // 未使用の空行は送らない(検証で除外済みだが、送信でも念のため除く)。
    cooking: values.ingredients
      .filter((m) => !isEmptyRow(m))
      .map((m) => ({
        ingredients: { name: m.name },
        quantity: m.quantity,
        unit: m.unit,
      })),
    season: values.seasoning
      .filter((m) => !isEmptyRow(m))
      .map((m) => ({
        seasoning: { name: m.name },
        quantity: m.quantity,
        unit: m.unit,
      })),
  }
}
