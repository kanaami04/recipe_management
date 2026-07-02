import { useQuery } from '@tanstack/react-query'

import { RecipeCardDialog } from '@/features/recipes/components/RecipeCardDialog'
import { RecipesHeader } from '@/features/recipes/components/RecipesHeader'
import { listRecipesOptions } from '@/shared/api/generated/@tanstack/react-query.gen'

export function RecipesPage() {
  // サーバ状態は生成 Query フックで取得する。認証は interceptor が付与する。
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
        <div className="grid grid-cols-2 gap-3 p-3 sm:grid-cols-3 sm:gap-4 sm:p-4 lg:grid-cols-4 xl:grid-cols-5">
          {recipesData.map((recipe) => (
            <RecipeCardDialog key={recipe.id} recipe={recipe} />
          ))}
        </div>
      )}
    </>
  )
}
