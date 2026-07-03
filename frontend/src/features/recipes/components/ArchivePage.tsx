import { useQuery } from '@tanstack/react-query'

import { RecipeCardDialog } from '@/features/recipes/components/RecipeCardDialog'
import { listRecipesOptions } from '@/shared/api/generated/@tanstack/react-query.gen'
import { Separator } from '@/shared/ui/separator'
import { SidebarTrigger } from '@/shared/ui/sidebar'

// アーカイブ一覧。取得済みのレシピからアーカイブ済みだけを表示する
// (メイン一覧は非アーカイブのみ)。並び替えは持たない。
export function ArchivePage() {
  const { data: recipesData, isPending, isError } = useQuery(listRecipesOptions())
  const archived = recipesData?.filter((r) => r.archive_flg) ?? []

  return (
    <>
      <header className="flex h-(--header-height) shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear">
        <div className="flex w-full items-center gap-2 px-3 sm:px-4 lg:px-6">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mx-1 data-[orientation=vertical]:h-4" />
          <h1 className="text-base font-medium">アーカイブ</h1>
        </div>
      </header>
      {isPending ? (
        <div className="flex items-center justify-center min-h-60">
          <p className="text-muted-foreground">読み込み中...</p>
        </div>
      ) : isError ? (
        <div className="flex items-center justify-center min-h-60">
          <p className="text-destructive">レシピの取得に失敗しました</p>
        </div>
      ) : archived.length === 0 ? (
        <div className="flex items-center justify-center min-h-60">
          <p className="text-muted-foreground">アーカイブされたレシピはありません</p>
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-3 p-3 sm:grid-cols-3 sm:gap-4 sm:p-4 lg:grid-cols-4 xl:grid-cols-5">
          {archived.map((recipe) => (
            <RecipeCardDialog key={recipe.id} recipe={recipe} />
          ))}
        </div>
      )}
    </>
  )
}
