import { Button } from "../ui/button"
import { Separator } from "@/components/ui/separator"
import { SidebarTrigger } from "@/components/ui/sidebar"
import { Input } from "@/components/ui/input"
import { RecipeForm } from "@/components/recipes/RecipeForm"
import { api } from "@/lib/api"
import { useUser } from "@/hooks/UserContext"
import { useFetchRecipeLabel, useFetchSharedUser } from "@/hooks/useFetchData"
import { mutate } from "swr"
import {
	Dialog,
	DialogTrigger,
	DialogTitle,
	DialogHeader,
	DialogDescription,
	DialogContent,
} from "@/components/ui/dialog";
import type { RecipeDataType } from "@/type/RecipeDataType"
import { useState } from "react"
import axios from "axios"
import { MessageAlertDialog } from "@/components/MessageAlertDialog";

export function RecipesHeader() {

	const { token } = useUser()
	const { data: sharedUserData } = useFetchSharedUser(token)
	const { data: labelData } = useFetchRecipeLabel(token)

	const [isOpen, setIsOpen] = useState(false)
	const [isSuccessOpen, setIsSuccessOpen] = useState(false)

	const handleCreate = async (payload: RecipeDataType) => {
		if (!token) {
			alert("作成できません。ログインしてください。")
			return
		}
		try {
			const res = await api.post('/api/recipes/', payload, {
				headers: { Authorization: `Bearer ${token}` },
			})

			console.log("レシピ作成成功", res.data)
			setIsSuccessOpen(true)
		} catch (error: unknown) {
			console.error("レシピ作成失敗", error)
			if (axios.isAxiosError(error)) {
				console.error("APIエラー詳細:", error.response?.data);
				alert("作成に失敗しました: " + JSON.stringify(error.response?.data));
			} else {
				alert("通信エラーが発生しました。");
			}
		}

		mutate("/api/recipes/")
		setIsOpen(false)
	}

	return (
		<header className="flex h-(--header-height) shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-(--header-height)">
			<div className="flex w-full items-center gap-1 lg:gap-2 lg:px-6">
				<SidebarTrigger className="-ml-1" />
				<Separator
					orientation="vertical"
					className="mx-1 data-[orientation=vertical]:h-4"
				/>
				<h1 className="text-base font-medium">MyRecipes</h1>
				<div className="ml-auto flex items-center gap-2">
					<Input placeholder="search" className="sm-w-40" />
					<Separator
						orientation="vertical"
						className="data-[orientation=vertical]:h-4"
					/>
					<Dialog open={isOpen} onOpenChange={setIsOpen}>
						<DialogTrigger asChild>
							<Button size="sm" onClick={() => setIsOpen(true)}> + </Button>
						</DialogTrigger>
						<DialogContent className="max-w-3xl w-full">
							<DialogHeader>
								<DialogTitle>new recipe</DialogTitle>
								<DialogDescription>create new recipe.</DialogDescription>
							</DialogHeader>
							<RecipeForm
								mode='create'
								onSubmit={handleCreate}
								labelData={labelData}
								sharedUserData={sharedUserData}
							/>
						</DialogContent>
					</Dialog >
					<MessageAlertDialog
						title="レシピ作成成功"
						description={`レシピが作成されました。`}
						open={isSuccessOpen}
						onOpenChange={() => setIsSuccessOpen(false)}
					/>
				</div>
			</div>
		</header>
	)
}