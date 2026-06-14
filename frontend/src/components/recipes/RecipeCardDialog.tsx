import type { RecipeDataType } from "@/type/RecipeDataType";
import {
	Dialog,
	DialogTrigger,
	DialogTitle,
	DialogHeader,
	DialogFooter,
	DialogClose,
	DialogContent,
} from "@/components/ui/dialog";
import { RecipeCard } from "./RecipeCard";
import { Button } from "../ui/button";
import { DialogDescription } from "@radix-ui/react-dialog";
import { useState } from "react";
import { RecipeDetailEditDialog } from "./RecipeDetailEditDialog";
import { Label } from "../ui/label";
import { useUser } from "@/hooks/UserContext";
import { api } from "@/lib/api";
import axios from "axios";
import { mutate } from "swr"

export function RecipeCardDialog({ recipe }: { recipe: RecipeDataType }) {
	const [isEditing, setIsEditing] = useState(false)
	const [isOpen, setIsOpen] = useState(false)

	const { token } = useUser()

	const handleDeleteRecipe = async () => {
		if (!token) {
			return
		}

		try {
			const res = await api.delete(`/api/recipes/${recipe.id}/`, {
				headers: { Authorization: `Bearer ${token}` },
			})
			console.log("レシピ削除成功", res.data)

		} catch (error: unknown) {
			console.error("レシピ削除失敗", error)
			if (axios.isAxiosError(error)) {
				console.error("APIエラー詳細:", error.response?.data);
				alert("削除に失敗しました: " + JSON.stringify(error.response?.data));
			} else {
				alert("通信エラーが発生しました。");
			}
		}
		mutate("/api/recipes/")
		setIsOpen(false);
	}


	return (
		<>
			<Dialog open={isOpen} onOpenChange={setIsOpen}>
				<DialogTrigger asChild>
					<div role="button" style={{ display: 'inline-block', cursor: 'pointer' }}>
						<RecipeCard key={recipe.id} recipe={recipe} />
					</div>
				</DialogTrigger>
				<DialogContent className="max-w-3xl w-full">
					<DialogHeader>
						<DialogTitle>{recipe.title}</DialogTitle>
						<DialogDescription>レシピの詳細</DialogDescription>
					</DialogHeader>
					<div className="gap-4 h-140 overflow-auto">
						<div className="grid gap-3">
							<div className="flex gap-3 w-full justify-start">
								<Label>{recipe.create_for}人前</Label>
								<Label>{recipe.create_time}分</Label>
							</div>
							<div className="flex-2 gap-3">
								<Label>食材</Label>
								{recipe.cooking.map((c) =>
									<p key={c.ingredients.name}>
										{c.ingredients.name} : {c.quantity}{c.unit}
									</p>
								)}

								<Label>調味料</Label>
								{recipe.season.map((s) =>
									<p key={s.seasoning.name}>
										{s.seasoning.name} : {s.quantity}{s.unit}
									</p>
								)}
								<Label>説明</Label>
								<p>{recipe.procedure}</p>
							</div>
							<div>
								{recipe.label.map((label) =>
									<p key={label.name}>{label.name}</p>
								)}
							</div>
							<div>
								{recipe.shared_user.map((user) =>
									<p key={user.username}>{user.username}</p>
								)}
							</div>
						</div>
					</div>
					<DialogFooter>
						<DialogClose asChild>
							<Button variant="outline">Cancel</Button>
						</DialogClose>
						<Button type="button" onClick={() => setIsEditing(true)}>Edit</Button>
						<Button type="button" onClick={handleDeleteRecipe}>Delete</Button>
					</DialogFooter>
				</DialogContent>
				<RecipeDetailEditDialog recipe={recipe} open={isEditing} onOpenChange={() => setIsEditing(false)} />
			</Dialog>
		</>
	)
}