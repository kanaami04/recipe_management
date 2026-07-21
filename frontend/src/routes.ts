import { index, route, type RouteConfig } from '@react-router/dev/routes'

// 薄いルート定義。UI ロジックは feature/components 側へ委譲する。
// 既存 URL 構造(/ = ログイン、/top = 保護レイアウト + レシピ一覧)を維持する。
export default [
  index('routes/login.tsx'),
  route('signup', 'routes/signup.tsx'),
  // メール確認・パスワードリセットは未ログインで到達するため公開ルートに置く。
  route('verify-email', 'routes/verify-email.tsx'),
  route('reset-password', 'routes/reset-password.tsx'),
  route('reset-password/confirm', 'routes/reset-password-confirm.tsx'),
  route('top', 'routes/protected.tsx', [
    index('routes/recipes.tsx'),
    route('shopping-list', 'routes/shopping-list.tsx'),
    route('archive', 'routes/archive.tsx'),
    route('labels', 'routes/labels.tsx'),
    route('share-group', 'routes/share-group.tsx'),
    route('account', 'routes/account.tsx'),
    route('account/email', 'routes/account-email.tsx'),
  ]),
  route('*', 'routes/catchall.tsx'),
] satisfies RouteConfig
