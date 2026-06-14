# 0006. フォームとバリデーション: React Hook Form + Zod

- ステータス: Accepted
- 日付: 2026-06-14

## コンテキスト

`RecipeForm` は useState を 10 個手動管理し、バリデーションが一切ない。空タイトルでも送信でき、
`Number(time)` が `NaN` になりうる。送信時には payload を手で組み立て、API 形状
(`cooking: {ingredients:{name}}` 等)への変換ロジックがコンポーネントに埋まっている。
`LoginForm` も同様に手動 useState で、`<form>` の構造にもバグがある。

宣言的な検証と、型・初期値・エラー表示の一元管理が必要。

## 決定

### 1. React Hook Form + Zod + zodResolver を採用する

`useForm({ resolver: zodResolver(schema) })` でフォーム状態を管理し、手動 useState 群を撤廃する。
shadcn の `Form` コンポーネント(RHF 公式連携)を導入し、エラー表示を `FormMessage` で標準化する
(現状未導入のため追加する)。

### 2. zod スキーマをフォーム型の単一の源にする

```ts
const recipeFormSchema = z.object({
  title: z.string().min(1, "タイトルは必須です"),
  createTime: z.coerce.number().int().positive(), // 文字列→数値の変換と検証を一元化
  ingredients: z.array(materialSchema).min(1),
  // ...
})
type RecipeFormValues = z.infer<typeof recipeFormSchema>
```

型・初期値(`defaultValues`)・検証・エラー文言をスキーマ 1 つに集約する。

### 3. 「フォームの形」と「API の形」を分離する

フォーム用スキーマは UI に素直な形にし、送信時に API 形状へ変換する関数を `features/*/api` 側へ置く。
現状コンポーネントに埋まっている変換ロジックをそこへ追い出す。

> フォーム用 zod は手書きする。API レスポンス用 zod は OpenAPI から生成し、住み分ける
> ([ADR-0007](0007-api-type-sync-openapi.md))。

### 4. 検証タイミング

`onSubmit` + `onBlur` を基本とし、過剰な onChange 検証は避ける。

## 結果

### 良い点

- 必須・数値・件数などの検証が宣言的になり、不正な送信を防げる。
- 型・初期値・検証・エラー表示が 1 スキーマに集約される。
- `create`/`edit` の初期値注入(現状 `?? ''` の山)が `defaultValues` で整理される。

### トレードオフ

- RHF / Controller と shadcn コンポーネントの接続を学ぶコストがある。
- フォーム型と API 型の変換層を別途設ける必要がある(意図的な分離)。
