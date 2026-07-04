import { index, route, type RouteConfig } from '@react-router/dev/routes'

// 薄いルート定義。UI ロジックは feature/components 側へ委譲する。
// 既存 URL 構造(/ = ログイン、/top = 保護レイアウト + レシピ一覧)を維持する。
export default [
  index('routes/login.tsx'),
  route('signup', 'routes/signup.tsx'),
  route('top', 'routes/protected.tsx', [
    index('routes/recipes.tsx'),
    route('archive', 'routes/archive.tsx'),
    route('labels', 'routes/labels.tsx'),
  ]),
  route('*', 'routes/catchall.tsx'),
] satisfies RouteConfig
