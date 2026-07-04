import { useQueryClient } from '@tanstack/react-query'

import { logout } from '@/shared/auth/authClient'
import { useUser } from '@/shared/auth/UserContext'
import { useNavigateAndClose } from '@/shared/components/sidebar/useNavigateAndClose'
import { Avatar, AvatarFallback, AvatarImage } from '@/shared/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/shared/ui/dropdown-menu'
import { SidebarMenu, SidebarMenuButton, SidebarMenuItem, useSidebar } from '@/shared/ui/sidebar'

export function NavUser() {
  const { isMobile } = useSidebar()
  const { user } = useUser()
  const queryClient = useQueryClient()
  const navigateAndClose = useNavigateAndClose()

  const onClickAccount = () => navigateAndClose('/top/account')

  const onClickLogout = async () => {
    // サーバ側で refresh Cookie を失効し、メモリの access を消す。
    await logout()
    // ユーザー固有のキャッシュを破棄し、別アカウントで再ログインしたとき
    // 前ユーザーの情報が一瞬残らないようにする。
    queryClient.clear()
    navigateAndClose('/')
  }

  const onClickLogIn = () => navigateAndClose('/')

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <Avatar>
                <AvatarImage src={user?.avatar_url ?? undefined} alt="プロフィール画像" />
                <AvatarFallback className="rounded-lg">
                  {user ? user.username.charAt(0).toUpperCase() : 'U'}
                </AvatarFallback>
              </Avatar>
              <div className="grid flex-1 text-left text-sm leading-tight">
                {user ? (
                  <>
                    <span className="truncate font-medium">{user.username}</span>
                    <span className="text-muted-foreground truncate text-xs">{user.email}</span>
                  </>
                ) : (
                  <span className="truncate font-medium">no login</span>
                )}
              </div>
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
            side={isMobile ? 'bottom' : 'right'}
            align="end"
            sideOffset={4}
          >
            <DropdownMenuLabel className="p-0 font-normal">
              <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                <Avatar className="h-8 w-8 rounded-lg">
                  <AvatarImage src={user?.avatar_url ?? undefined} alt="プロフィール画像" />
                  <AvatarFallback className="rounded-lg">
                    {user ? user.username.charAt(0).toUpperCase() : 'U'}
                  </AvatarFallback>
                </Avatar>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  {user ? (
                    <>
                      <span className="truncate font-medium">{user.username}</span>
                      <span className="text-muted-foreground truncate text-xs">{user.email}</span>
                    </>
                  ) : (
                    <span className="truncate font-medium">未ログイン</span>
                  )}
                </div>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            {user ? (
              <>
                <DropdownMenuItem onClick={onClickAccount}>アカウント</DropdownMenuItem>
                <DropdownMenuItem onClick={onClickLogout}>ログアウト</DropdownMenuItem>
              </>
            ) : (
              <DropdownMenuItem onClick={onClickLogIn}>ログイン</DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
