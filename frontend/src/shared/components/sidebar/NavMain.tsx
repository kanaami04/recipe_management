import { useNavigate } from 'react-router-dom'

import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/shared/ui/sidebar'

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
              tooltip="レシピ一覧"
              className="bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground active:bg-primary/90 active:text-primary-foreground min-w-8 duration-200 ease-linear"
              onClick={onClickCreateRecipe}
            >
              <span>レシピ</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          {/* ラベル一覧画面へ遷移するボタン */}
          <SidebarMenuItem className="flex items-center gap-2">
            <SidebarMenuButton
              tooltip="ラベル一覧"
              className="bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground active:bg-primary/90 active:text-primary-foreground min-w-8 duration-200 ease-linear"
              onClick={onClickCreateLabel}
            >
              <span>ラベル</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
        <SidebarMenu>
          {/* アーカイブ一覧画面へ遷移するボタン */}
          <SidebarMenuItem>
            <SidebarMenuButton onClick={onClickArchive}>
              <span>アーカイブ</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}
