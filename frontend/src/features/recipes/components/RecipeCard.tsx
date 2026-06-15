import type { RecipeResponse } from '@/shared/api/generated/types.gen'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { Label } from '@/shared/ui/label'

export function RecipeCard({ recipe }: { recipe: RecipeResponse }) {
  return (
    <Card className="h-full w-full gap-2 transition-shadow hover:shadow-md">
      <CardHeader className="text-center">
        <CardTitle className="truncate">{recipe.title}</CardTitle>
      </CardHeader>
      <CardContent className="text-center text-sm">
        <div className="my-2">
          <Label className="justify-center">食材</Label>
          {recipe.cooking.map((cooking) => (
            <p key={cooking.ingredients.name} className="truncate">
              {cooking.ingredients.name} : {cooking.quantity}
              {cooking.unit}
            </p>
          ))}
        </div>
        <Label className="justify-center">調味料</Label>
        {recipe.season.map((season) => (
          <p key={season.seasoning.name} className="truncate">
            {season.seasoning.name} : {season.quantity}
            {season.unit}
          </p>
        ))}
      </CardContent>
    </Card>
  )
}
