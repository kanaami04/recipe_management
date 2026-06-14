import {
	SidebarGroup,
	SidebarGroupContent,
	SidebarMenu,
	SidebarMenuItem,
	SidebarMenuButton
} from "@/components/ui/sidebar";
import { useNavigate } from "react-router-dom";

export function NavMain() {
	const navigate = useNavigate()

	const onClickCreateRecipe = () => {
		navigate('/top')
	}

	const onClickCreateLabel = () => {
		navigate('/top/createLabel')
	}

	const onClickArchive = () => {
		navigate('/top/archive')
	}

	return (
		<SidebarGroup>
			<SidebarGroupContent className="flex flex-col gap-2">
				<SidebarMenu>
					{/* レシピ一覧画面へ遷移するボタン */}
					<SidebarMenuItem className="flex items-center gap-2">
						<SidebarMenuButton
							tooltip="Quick Create"
							className="bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground active:bg-primary/90 active:text-primary-foreground min-w-8 duration-200 ease-linear"
							onClick={onClickCreateRecipe}
						>
							<span>Recipe</span>
						</SidebarMenuButton>
					</SidebarMenuItem>
					{/* ラベル一覧画面へ遷移するボタン */}
					<SidebarMenuItem className="flex items-center gap-2">
						<SidebarMenuButton
							tooltip="Quick Create"
							className="bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground active:bg-primary/90 active:text-primary-foreground min-w-8 duration-200 ease-linear"
							onClick={onClickCreateLabel}
						>
							<span>Label</span>
						</SidebarMenuButton>
					</SidebarMenuItem>
				</SidebarMenu>
				<SidebarMenu>
					{/* アーカイブ一覧画面へ遷移するボタン */}
					<SidebarMenuItem>
						<SidebarMenuButton onClick={onClickArchive}>
							<span>Archive</span>
						</SidebarMenuButton>
					</SidebarMenuItem>
				</SidebarMenu>
			</SidebarGroupContent>
		</SidebarGroup>
	)
}