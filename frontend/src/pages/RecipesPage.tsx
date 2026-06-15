import { useQuery } from '@tanstack/react-query'

import { RecipeCardDialog } from '@/components/recipes/RecipeCardDialog'
import { RecipesHeader } from '@/components/recipes/RecipesHeader'
import { listRecipesOptions } from '@/shared/api/generated/@tanstack/react-query.gen'

export function RecipesPage() {
  // サーバ状態は生成 Query フックで取得する (ADR-0003/0007)。認証は interceptor が付与する。
  const { data: recipesData, error } = useQuery(listRecipesOptions())

  return (
    <>
      <RecipesHeader />
      <div className="flex flex-wrap">
        {recipesData ? (
          recipesData.map((recipe) => (
            <div key={recipe.id} className="gap-2 py-2 md:gap-4 md:py-4 ">
              <RecipeCardDialog recipe={recipe} />
            </div>
          ))
        ) : (
          <div className="flex items-center justify-center min-h-screen">
            {error ? <p>fetch error</p> : <p>no data...</p>}
          </div>
        )}
      </div>
    </>
  )
}
