import { Outlet } from 'react-router'

import { AppSidebar } from '@/components/sidebar/AppSidebar'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { requireAuth } from '@/shared/auth/requireAuth'

// 認証ガードを clientLoader に集約する (ADR-0002 / ADR-0004)。
// access が無ければ Cookie の refresh で復帰を試み、失敗なら "/" へ。
export const clientLoader = requireAuth

// 保護レイアウト。認証は clientLoader で担保済み。
export default function ProtectedLayout() {
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
