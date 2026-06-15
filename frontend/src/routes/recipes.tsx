import { RecipesPage } from '@/pages/RecipesPage'

// 薄いルート: 描画は feature 側に委譲する (ADR-0002)。
export default function RecipesRoute() {
  return <RecipesPage />
}
