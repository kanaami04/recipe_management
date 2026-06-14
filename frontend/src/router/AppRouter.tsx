import { Routes, Route, Navigate, Outlet } from "react-router-dom";
import { LoginPage } from "@/pages/LoginPage";
import { AppSidebar } from "@/components/sidebar/AppSidebar.tsx";
import { RecipesPage } from "@/pages/RecipesPage.tsx";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { useUser } from "@/hooks/UserContext";
let isAuthenticated = false;

// 保護ルート
const ProtectedLayout = () => {
    const { token } = useUser()

    if (token) {
        isAuthenticated = true
    }

    if (!isAuthenticated) {
        return <Navigate to="/" replace />;
    }

    return (
        <SidebarProvider
            style={
                {
                    "--sidebar-width": "calc(var(--spacing) * 72)",
                    "--header-height": "calc(var(--spacing) * 12)",
                } as React.CSSProperties
            }
        >
            <AppSidebar variant="inset"/>
            <SidebarInset>
                <main className="h-full">
                    <Outlet />
                </main>
            </SidebarInset>
        </SidebarProvider>
    );
};

export const AppRouter = () => {
    return (
        <Routes>
            {/* ログインページ */}
            <Route path="/" element={<LoginPage />} />

            {/* ログイン後の保護されたレイアウト */}
            <Route path="/top" element={<ProtectedLayout />}>
                {/* この中に表示したいページをネスト */}
                <Route index element={<RecipesPage />} />
            </Route>

            {/* それ以外のパスはログインにリダイレクト */}
            <Route path="*" element={<Navigate to="/" />} />
        </Routes>
    );
};