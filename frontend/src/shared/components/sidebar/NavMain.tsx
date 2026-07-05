import { useLocation } from 'react-router-dom'

import { useNavigateAndClose } from '@/shared/components/sidebar/useNavigateAndClose'
import { cn } from '@/shared/lib/utils'
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/shared/ui/sidebar'

export function NavMain() {
  const { pathname } = useLocation()
  const navigateAndClose = useNavigateAndClose()

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
              onClick={() => navigateAndClose('/top')}
            >
              <span>レシピ</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          {/* 買い物リスト画面へ遷移するボタン */}
          <SidebarMenuItem className="flex items-center gap-2">
            <SidebarMenuButton
              tooltip="買い物リスト"
              className={cn(
                'min-w-8 duration-200 ease-linear',
                pathname === '/top/shopping-list' && activeClass,
              )}
              onClick={() => navigateAndClose('/top/shopping-list')}
            >
              <span>買い物リスト</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          {/* ラベル管理画面へ遷移するボタン */}
          <SidebarMenuItem className="flex items-center gap-2">
            <SidebarMenuButton
              tooltip="ラベル管理"
              className={cn(
                'min-w-8 duration-200 ease-linear',
                pathname === '/top/labels' && activeClass,
              )}
              onClick={() => navigateAndClose('/top/labels')}
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
              onClick={() => navigateAndClose('/top/archive')}
            >
              <span>アーカイブ</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          {/* 共有グループ管理画面へ遷移するボタン */}
          <SidebarMenuItem>
            <SidebarMenuButton
              tooltip="共有グループ"
              className={cn(pathname === '/top/share-group' && activeClass)}
              onClick={() => navigateAndClose('/top/share-group')}
            >
              <span>共有グループ</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}
