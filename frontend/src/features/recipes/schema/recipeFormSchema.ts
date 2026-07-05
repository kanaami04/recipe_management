import { z } from 'zod'

import type { RecipeRequest, RecipeResponse } from '@/shared/api/generated/types.gen'

// フォーム用の zod スキーマ(手書き)。UI に素直な形にする。
// API レスポンス用 zod は生成物を使い、住み分ける。
const materialSchema = z.object({
  name: z.string(),
  quantity: z.number().min(0),
  // 単位は候補チップのほか手入力(その他)も許す。DB が varchar(10) のため 10 文字上限。
  unit: z.string().max(10, '単位は10文字以内で入力してください'),
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

// 入力モード。manual = 手動で全項目を入力、url = 参考 URL から登録(必須項目なし)。
export const recipeInputModes = ['manual', 'url'] as const
export type RecipeInputMode = (typeof recipeInputModes)[number]

export const recipeFormSchema = z
  .object({
    inputMode: z.enum(recipeInputModes),
    title: z.string(),
    createFor: z.string(),
    createTime: z.string(), // 任意。空文字を許容し、変換時に null へ。
    procedure: z.string(),
    // 参考にした外部レシピの URL と、その OGP サムネイル画像 URL。
    sourceUrl: z.string(),
    thumbnailUrl: z.string(),
    label: z.array(z.string()),
    // 空行は無視しつつ、入力途中の行は名前・単位を必須にする(モード共通)。
    ingredients: z.array(materialSchema).superRefine(validateRows),
    seasoning: z.array(materialSchema).superRefine(validateRows),
  })
  // 必須条件はモードで切り替える。url モードでは URL のみ必須、手動入力の項目は問わない。
  .superRefine((values, ctx) => {
    if (values.inputMode === 'url') {
      const src = values.sourceUrl.trim()
      if (src === '') {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ['sourceUrl'],
          message: '参考レシピの URL を入力してください',
        })
      } else if (!/^https?:\/\//i.test(src)) {
        // http(s) 以外はサムネ取得もできず、保存しても正しくリンクできないため弾く。
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          path: ['sourceUrl'],
          message: 'http:// または https:// で始まる URL を入力してください',
        })
      }
      return
    }
    // manual モード: タイトル・人数・食材(完全な行が1つ以上)を必須にする。
    if (values.title.trim() === '') {
      ctx.addIssue({ code: z.ZodIssueCode.custom, path: ['title'], message: 'タイトルは必須です' })
    }
    if (values.createFor === '') {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['createFor'],
        message: '人数を選択してください',
      })
    }
    if (!values.ingredients.some(isCompleteRow)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['ingredients'],
        message: '食材は1つ以上必要です',
      })
    }
  })

export type RecipeFormValues = z.infer<typeof recipeFormSchema>

// API レスポンス(RecipeResponse)→ フォーム値。defaultValues に使う。
// inputMode は UI 都合の状態。未指定なら参考 URL の有無で初期モードを決める。
export function toFormValues(
  recipe?: RecipeResponse,
  inputMode?: RecipeInputMode,
): RecipeFormValues {
  return {
    inputMode: inputMode ?? (recipe?.source_url ? 'url' : 'manual'),
    title: recipe?.title ?? '',
    createFor: recipe ? String(recipe.create_for) : '',
    createTime: recipe?.create_time != null ? String(recipe.create_time) : '',
    procedure: recipe?.procedure ?? '',
    sourceUrl: recipe?.source_url ?? '',
    thumbnailUrl: recipe?.thumbnail_url ?? '',
    label: recipe?.label.map((l) => l.name) ?? [],
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

// タイトル未入力(url モードで OGP も取れなかった等)を URL のホスト名で補う。
// API はタイトル必須のため、一覧で識別できる名前を必ず用意する。
function resolveTitle(values: RecipeFormValues): string {
  if (values.title.trim() !== '') return values.title
  try {
    const host = new URL(values.sourceUrl).hostname.replace(/^www\./, '')
    return host !== '' ? `${host} のレシピ` : '無題のレシピ'
  } catch {
    return '無題のレシピ'
  }
}

// フォーム値 → API リクエスト(RecipeRequest)。id は URL 側で扱う。
// 送信内容はモードに揃える。手動モードでは参考 URL を持たせず、url モードでは
// 画面に無い手動入力の子データ(食材・調味料・手順)を持たせない。これにより
// モード切替で残ったフォーム状態が中途半端なデータとして保存されるのを防ぐ。
export function toRecipeRequest(values: RecipeFormValues): RecipeRequest {
  const isUrl = values.inputMode === 'url'
  return {
    title: resolveTitle(values),
    create_for: Number(values.createFor),
    create_time: values.createTime === '' ? null : Number(values.createTime),
    procedure: isUrl ? '' : values.procedure,
    source_url: isUrl ? values.sourceUrl : '',
    thumbnail_url: isUrl ? values.thumbnailUrl : '',
    label: values.label.map((name) => ({ name })),
    // 未使用の空行は送らない(検証で除外済みだが、送信でも念のため除く)。
    cooking: isUrl
      ? []
      : values.ingredients
          .filter((m) => !isEmptyRow(m))
          .map((m) => ({
            ingredients: { name: m.name },
            quantity: m.quantity,
            unit: m.unit,
          })),
    season: isUrl
      ? []
      : values.seasoning
          .filter((m) => !isEmptyRow(m))
          .map((m) => ({
            seasoning: { name: m.name },
            quantity: m.quantity,
            unit: m.unit,
          })),
  }
}
