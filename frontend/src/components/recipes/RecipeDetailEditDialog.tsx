import type { RecipeDataType } from "@/type/RecipeDataType"
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogDescription,
} from "@/components/ui/dialog"
import { mutate } from "swr";
import { useFetchRecipeLabel, useFetchSharedUser } from "@/hooks/useFetchData"
import { useUser } from "@/hooks/UserContext"
import { RecipeForm } from "./RecipeForm";
import { api } from "@/lib/api";
import axios from "axios";
import { useState } from "react";
import { MessageAlertDialog } from "@/components/MessageAlertDialog";

type EditDialog = {
	recipe: RecipeDataType;
	open: boolean;
	onOpenChange: () => void;
}

export function RecipeDetailEditDialog({ recipe, open, onOpenChange }: EditDialog) {
	const { token } = useUser()
	const { data: sharedUserData } = useFetchSharedUser(token)
	const { data: labelData } = useFetchRecipeLabel(token)
	const [isSuccessOpen, setIsSuccessOpen] = useState(false)
	

	const handleEdit = async (payload: RecipeDataType) => {
		if (!token) {
			alert("作成できません。ログインしてください。")
			return
		}
		try {
			const res = await api.put(`/api/recipes/${payload.id}/`, payload, {
				headers: { Authorization: `Bearer ${token}` },
			})

			console.log("レシピ編集成功", res.data)
			setIsSuccessOpen(true)
		} catch (error: unknown) {
			console.error("レシピ編集失敗", error)
			if (axios.isAxiosError(error)) {
				console.error("APIエラー詳細:", error.response?.data);
				alert("作成に失敗しました: " + JSON.stringify(error.response?.data));
			} else {
				alert("通信エラーが発生しました。");
			}
		}

		mutate("/api/recipes/")
		onOpenChange()
	}

	return (
		<>
			<Dialog open={open} onOpenChange={onOpenChange}>
				<DialogContent className="max-w-3xl w-full">
					<DialogHeader>
						<DialogTitle>レシピ編集</DialogTitle>
						<DialogDescription>レシピを編集します</DialogDescription>
					</DialogHeader>
					<RecipeForm
						mode="edit"
						initialData={recipe}
						onSubmit={handleEdit}
						labelData={labelData}
						sharedUserData={sharedUserData}
						onClickCancel={onOpenChange}
					/>
				</DialogContent>
			</Dialog >

			<MessageAlertDialog
				title="レシピ編集成功"
				description={`レシピが編集されました。`}
				open={isSuccessOpen}
				onOpenChange={() => setIsSuccessOpen(false)}
			/>
		</>
	)
}