import { Navigate, Outlet } from 'react-router'

import { AppSidebar } from '@/components/sidebar/AppSidebar'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { useUser } from '@/hooks/UserContext'

// 保護レイアウト。未ログインは "/" へリダイレクトする。
// 現状は token を直接見て判定する(グローバル可変変数バグを解消)。
// clientLoader による認証ガードへの集約は ADR-0004 のトークンストア導入時に行う。
export default function ProtectedLayout() {
  const { token } = useUser()

  if (!token) {
    return <Navigate to="/" replace />
  }

  return (
    <SidebarProvider
      style={
        {
          '--sidebar-width': 'calc(var(--spacing) * 72)',
          '--header-height': 'calc(var(--spacing) * 12)',
        } as React.CSSProperties
      }
    >
      <AppSidebar variant="inset" />
      <SidebarInset>
        <main className="h-full">
          <Outlet />
        </main>
      </SidebarInset>
    </SidebarProvider>
  )
}
