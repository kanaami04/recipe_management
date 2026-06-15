import { useNavigate } from 'react-router-dom'

import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@/components/ui/sidebar'
import { useUser } from '@/hooks/UserContext'
import { logout } from '@/shared/auth/authClient'

export function NavUser() {
  const { isMobile } = useSidebar()
  const { user } = useUser()
  const navigate = useNavigate()

  const onClickLogout = async () => {
    // サーバ側で refresh Cookie を失効し、メモリの access を消す (ADR-0004)。
    await logout()
    navigate('/')
  }

  const onClickLogIn = () => {
    navigate('/')
  }

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
                <AvatarImage src="/apple.png" alt="img" />
                <AvatarFallback className="rounded-lg">U</AvatarFallback>
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
                  <AvatarImage src="/apple.png" alt="img" />
                  <AvatarFallback className="rounded-lg">U</AvatarFallback>
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
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuItem>Account</DropdownMenuItem>
              <DropdownMenuItem>Billing</DropdownMenuItem>
              <DropdownMenuItem>Notifications</DropdownMenuItem>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            {user ? (
              <DropdownMenuItem onClick={onClickLogout}>Log out</DropdownMenuItem>
            ) : (
              <DropdownMenuItem onClick={onClickLogIn}>Log In</DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
