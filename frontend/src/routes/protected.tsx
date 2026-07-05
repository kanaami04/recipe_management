import { useEffect, useRef } from 'react'
import { Outlet, useLocation } from 'react-router'

import { requireAuth } from '@/shared/auth/requireAuth'
import { AppSidebar } from '@/shared/components/sidebar/AppSidebar'
import { SidebarInset, SidebarProvider } from '@/shared/ui/sidebar'

// 認証ガードを clientLoader に集約する。
// access が無ければ Cookie の refresh で復帰を試み、失敗なら "/" へ。
export const clientLoader = requireAuth

// 保護レイアウト。認証は clientLoader で担保済み。
export default function ProtectedLayout() {
  // スクロールは main が担う(window ではない)ため React Router の ScrollRestoration が
  // 効かない。main はレイアウト維持で子ルート跨ぎでも同一ノードのため、遷移で先頭へ戻す。
  const mainRef = useRef<HTMLElement>(null)
  const { pathname } = useLocation()
  useEffect(() => {
    mainRef.current?.scrollTo(0, 0)
  }, [pathname])

  return (
    <SidebarProvider
      className="h-svh"
      style={
        {
          '--sidebar-width': 'calc(var(--spacing) * 72)',
          '--header-height': 'calc(var(--spacing) * 12)',
        } as React.CSSProperties
      }
    >
      <AppSidebar variant="inset" />
      <SidebarInset className="overflow-hidden">
        <main ref={mainRef} className="h-full overflow-y-auto overscroll-none">
          <Outlet />
        </main>
      </SidebarInset>
    </SidebarProvider>
  )
}
