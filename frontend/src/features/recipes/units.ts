// 食材・調味料の単位と、単位ごとの数量の基準を定義する。
// 一般的な和食レシピの分量感に合わせ、単位で「開始値・刻み・範囲」を変える。

export type UnitConfig = {
  unit: string
  // false の単位(適量・少々)は数量を持たない。数量欄を出さず、表示も単位のみ。
  hasQuantity: boolean
  // 単位を選んだときの既定の数量。
  start: number
  // 数量候補の刻み。
  step: number
  min: number
  max: number
  // 大さじ・小さじ・カップは数量の前に単位を置く(例: 大さじ1)。他は後ろ(例: 100g)。
  before?: boolean
}

const q = (unit: string, start: number, step: number, min: number, max: number, before = false): UnitConfig => ({
  unit,
  hasQuantity: true,
  start,
  step,
  min,
  max,
  before,
})

const noQuantity = (unit: string): UnitConfig => ({
  unit,
  hasQuantity: false,
  start: 0,
  step: 0,
  min: 0,
  max: 0,
})

// g / ml / cc は 100 から 10 刻み。
const weight = (unit: string) => q(unit, 100, 10, 10, 1000)
// 個数系は 1 から 1 刻み。
const count = (unit: string) => q(unit, 1, 1, 1, 20)

export const INGREDIENT_UNITS: UnitConfig[] = [
  count('個'),
  count('本'),
  count('枚'),
  count('片'),
  count('束'),
  count('玉'),
  count('パック'),
  count('缶'),
  count('袋'),
  weight('g'),
  weight('ml'),
  noQuantity('適量'),
  noQuantity('少々'),
]

export const SEASONING_UNITS: UnitConfig[] = [
  q('大さじ', 1, 0.5, 0.5, 10, true),
  q('小さじ', 1, 0.5, 0.5, 10, true),
  q('カップ', 1, 0.5, 0.5, 5, true),
  weight('g'),
  weight('ml'),
  weight('cc'),
  noQuantity('少々'),
  noQuantity('適量'),
]

// 全単位を横断で引くための索引(表示整形などで単位名から config を得る)。
const UNIT_INDEX: Record<string, UnitConfig> = Object.fromEntries(
  [...INGREDIENT_UNITS, ...SEASONING_UNITS].map((u) => [u.unit, u]),
)

export function findUnit(unit: string): UnitConfig | undefined {
  return UNIT_INDEX[unit]
}

// 数量を分数付きの短い表記にする(0.5→"½"、1.5→"1½")。0.5 以外の端数はそのまま数値表記。
export function formatQuantity(quantity: number): string {
  if (Number.isInteger(quantity)) return String(quantity)
  const whole = Math.floor(quantity)
  const frac = quantity - whole
  if (frac === 0.5) return whole === 0 ? '½' : `${whole}½`
  return String(quantity)
}

// 単位に応じた数量候補。数量なし単位では空配列。
export function quantityOptions(config: UnitConfig): { label: string; value: string }[] {
  if (!config.hasQuantity) return []
  const options: { label: string; value: string }[] = []
  // 浮動小数の誤差を避けるため 2 桁で丸めてから比較・格納する。
  const round = (n: number) => Math.round(n * 100) / 100
  for (let v = config.min; v <= config.max + 1e-9; v = round(v + config.step)) {
    options.push({ label: formatQuantity(v), value: String(v) })
  }
  return options
}

// カード等での「数量 + 単位」の表示。適量・少々は単位のみ、大さじ系は単位を前に置く。
export function formatAmount(quantity: number, unit: string): string {
  const config = findUnit(unit)
  if (config && !config.hasQuantity) return unit
  if (config?.before) return `${unit}${formatQuantity(quantity)}`
  return `${formatQuantity(quantity)}${unit}`
}
