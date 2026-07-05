import type { RecipeResponse } from '@/shared/api/generated/types.gen'

import { RecipeSummaryCard } from './RecipeSummaryCard'
import { RecipeThumbnailCard } from './RecipeThumbnailCard'

// 一覧カードは登録方法で出し分ける。外部 URL から登録したレシピ(source_url あり)は
// サムネ主体のパネル、手動入力のレシピは従来の要約カードで表示する。
export function RecipeCard({ recipe }: { recipe: RecipeResponse }) {
  if (recipe.source_url !== '') {
    return <RecipeThumbnailCard recipe={recipe} />
  }
  return <RecipeSummaryCard recipe={recipe} />
}
