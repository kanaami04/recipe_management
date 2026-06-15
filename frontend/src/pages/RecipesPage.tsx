import { RecipeCardDialog } from '@/components/recipes/RecipeCardDialog'
import { RecipesHeader } from '@/components/recipes/RecipesHeader'
import { useFetchRecipes } from '@/hooks/useFetchData'
import { useUser } from '@/hooks/UserContext'

export function RecipesPage() {
  const { token } = useUser()

  const { data: recipesData, error } = useFetchRecipes(token)

  console.log('recipe', recipesData)

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
