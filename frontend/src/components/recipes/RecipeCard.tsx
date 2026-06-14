import { Card, CardContent, CardHeader, CardTitle } from "../ui/card";
import { Label } from "../ui/label";
import type { RecipeDataType } from "@/type/RecipeDataType";

export function RecipeCard({ recipe }: { recipe: RecipeDataType }) {
	return (
		<>
			<div className="*:data-[slot=card]:from-primary/5 *:data-[slot=card]:to-card dark:*:data-[slot=card]:bg-card grid grid-cols-1 gap-4 px-4 *:data-[slot=card]:bg-gradient-to-t *:data-[slot=card]:shadow-xs lg:px-6 @xl/main:grid-cols-2 @5xl/main:grid-cols-4">
				<Card className="@container/card w-50 h-60" >
					<CardHeader className="text-center">
						<CardTitle>{recipe.title}</CardTitle>
					</CardHeader>
					<CardContent>
						<div className="text-center">
							<div className="my-2">
								<Label>食材</Label>
								{recipe.cooking.map((cooking) =>
									<p key={recipe.id}>{cooking.ingredients.name} : {cooking.quantity}{cooking.unit}</p>
								)}
							</div>
							<Label>調味料</Label>
							{recipe.season.map((season) =>
								<p key={recipe.id}>{season.seasoning.name} : {season.unit}{season.quantity}</p>
							)}
						</div>
					</CardContent>
				</Card>
			</div>
		</>
	)
}