import { CookingPot } from 'lucide-react'

import type { RecipeResponse } from '@/shared/api/generated/types.gen'
import { Card, CardTitle } from '@/shared/ui/card'

import { RECIPE_CARD_HEIGHT_CLASS } from './recipeCardHeight'
import { RecipeSharedAvatars } from './RecipeSharedAvatars'

// 外部 URL から登録したレシピ用の一覧カード。上 8 割にサムネイル
// (OGP 画像が無ければ調理アイコンのプレースホルダ)、下 2 割にタイトルを置く。
export function RecipeThumbnailCard({ recipe }: { recipe: RecipeResponse }) {
  const hasThumbnail = recipe.thumbnail_url !== ''

  return (
    // 列幅で伸びる aspect ratio は使わず固定高にし、内部を 8:2(サムネ:タイトル)で分ける。
    <Card
      className={`${RECIPE_CARD_HEIGHT_CLASS} w-full gap-0 overflow-hidden py-0 transition-shadow hover:shadow-md`}
    >
      {/* サムネ領域(8 割)。画像が無いときは調理アイコンのプレースホルダで揃える。 */}
      <div className="relative flex-[4] overflow-hidden">
        {hasThumbnail ? (
          <img
            src={recipe.thumbnail_url}
            alt={recipe.title}
            className="size-full object-cover"
            loading="lazy"
          />
        ) : (
          <div className="bg-muted flex size-full items-center justify-center">
            <CookingPot className="text-muted-foreground/60 size-8" />
          </div>
        )}
        {/* 共有相手アバターはサムネ左下にオーバーレイして重ねる。 */}
        <RecipeSharedAvatars users={recipe.shared_user} className="absolute bottom-1.5 left-1.5" />
      </div>

      {/* タイトル領域(2 割)。要約カードと同じく 1 行の太字で、長いときは末尾を … で省略する。 */}
      <div className="flex flex-[1] items-center justify-center px-2 py-1">
        <CardTitle className="w-full truncate text-center text-sm font-semibold">
          {recipe.title}
        </CardTitle>
      </div>
    </Card>
  )
}
