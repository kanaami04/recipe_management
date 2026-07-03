import { useLocation, useNavigate } from 'react-router-dom'

import { cn } from '@/shared/lib/utils'
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/shared/ui/sidebar'

export function NavMain() {
  const navigate = useNavigate()
  const { pathname } = useLocation()

  // 現在のパスに応じてアクティブなナビを塗る。
  const activeClass =
    'bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground active:bg-primary/90 active:text-primary-foreground'

  return (
    <SidebarGroup>
      <SidebarGroupContent className="flex flex-col gap-2">
        <SidebarMenu>
          {/* レシピ一覧画面へ遷移するボタン */}
          <SidebarMenuItem className="flex items-center gap-2">
            <SidebarMenuButton
              tooltip="レシピ一覧"
              className={cn('min-w-8 duration-200 ease-linear', pathname === '/top' && activeClass)}
              onClick={() => navigate('/top')}
            >
              <span>レシピ</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          {/* ラベル一覧画面へ遷移するボタン */}
          <SidebarMenuItem className="flex items-center gap-2">
            <SidebarMenuButton
              tooltip="ラベル一覧"
              className="min-w-8 duration-200 ease-linear"
              onClick={() => navigate('/top/createLabel')}
            >
              <span>ラベル</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
        <SidebarMenu>
          {/* アーカイブ一覧画面へ遷移するボタン */}
          <SidebarMenuItem>
            <SidebarMenuButton
              tooltip="アーカイブ一覧"
              className={cn(pathname === '/top/archive' && activeClass)}
              onClick={() => navigate('/top/archive')}
            >
              <span>アーカイブ</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}
