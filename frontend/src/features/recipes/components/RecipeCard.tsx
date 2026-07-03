import { Clock, Users } from 'lucide-react'

import { splitSteps } from '@/features/recipes/steps'
import type { RecipeResponse } from '@/shared/api/generated/types.gen'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'

// 一覧カードは「要約プレビュー」。全リストは載せず、人前・時間(アイコン)と
// 食材/調味料/手順の件数だけを出す。詳細はタップで開く(RecipeCardDialog)。
export function RecipeCard({ recipe }: { recipe: RecipeResponse }) {
  const stepCount = splitSteps(recipe.procedure).filter((s) => s.trim() !== '').length

  return (
    <Card className="h-full w-full gap-2 transition-shadow hover:shadow-md">
      <CardHeader>
        <CardTitle className="truncate text-center">{recipe.title}</CardTitle>
      </CardHeader>
      <CardContent className="text-muted-foreground flex flex-col items-center gap-1.5 text-sm">
        <div className="flex items-center gap-3">
          <span className="flex items-center gap-1">
            <Users className="size-4" />
            {recipe.create_for}
          </span>
          {recipe.create_time != null && (
            <span className="flex items-center gap-1">
              <Clock className="size-4" />
              {recipe.create_time}分
            </span>
          )}
        </div>
        <div className="text-xs">
          食材{recipe.cooking.length}・調味料{recipe.season.length}・手順{stepCount}
        </div>
      </CardContent>
    </Card>
  )
}
