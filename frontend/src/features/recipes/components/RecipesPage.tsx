import { useQuery } from '@tanstack/react-query'

import { RecipeCardDialog } from '@/features/recipes/components/RecipeCardDialog'
import { RecipesHeader } from '@/features/recipes/components/RecipesHeader'
import { listRecipesOptions } from '@/shared/api/generated/@tanstack/react-query.gen'

export function RecipesPage() {
  // サーバ状態は生成 Query フックで取得する (ADR-0003/0007)。認証は interceptor が付与する。
  const { data: recipesData, isPending, isError } = useQuery(listRecipesOptions())

  return (
    <>
      <RecipesHeader />
      {isPending ? (
        <div className="flex items-center justify-center min-h-60">
          <p className="text-muted-foreground">読み込み中...</p>
        </div>
      ) : isError ? (
        <div className="flex items-center justify-center min-h-60">
          <p className="text-destructive">レシピの取得に失敗しました</p>
        </div>
      ) : recipesData.length === 0 ? (
        <div className="flex items-center justify-center min-h-60">
          <p className="text-muted-foreground">レシピがまだありません</p>
        </div>
      ) : (
        <div className="flex flex-wrap">
          {recipesData.map((recipe) => (
            <div key={recipe.id} className="gap-2 py-2 md:gap-4 md:py-4 ">
              <RecipeCardDialog recipe={recipe} />
            </div>
          ))}
        </div>
      )}
    </>
  )
}
