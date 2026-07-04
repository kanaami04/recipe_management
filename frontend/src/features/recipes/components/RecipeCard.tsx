import { Clock, Users } from 'lucide-react'

import { splitSteps } from '@/features/recipes/steps'
import type { RecipeResponse, UserListItem } from '@/shared/api/generated/types.gen'
import { cn } from '@/shared/lib/utils'
import { Avatar, AvatarFallback, AvatarImage } from '@/shared/ui/avatar'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'

// 一覧カードに並べる共有相手アバターの最大数。超過分は「+N」でまとめる
// (カード幅・高さを共有人数に依らず一定に保つため)。
const MAX_AVATARS = 3

// 一覧カードは「要約プレビュー」。全リストは載せず、人前・時間(アイコン)と
// 食材/調味料/手順の件数だけを出す。詳細はタップで開く(RecipeCardDialog)。
export function RecipeCard({ recipe }: { recipe: RecipeResponse }) {
  const stepCount = splitSteps(recipe.procedure).filter((s) => s.trim() !== '').length

  return (
    <Card className="h-full w-full gap-2 transition-shadow hover:shadow-md">
      <CardHeader>
        <CardTitle className="truncate text-center">{recipe.title}</CardTitle>
      </CardHeader>
      <CardContent className="text-muted-foreground flex flex-col items-center gap-1.5 text-sm">
        <div className="flex items-center gap-3">
          <span className="flex items-center gap-1">
            <Users className="size-4" />
            {recipe.create_for}
          </span>
          {recipe.create_time != null && (
            <span className="flex items-center gap-1">
              <Clock className="size-4" />
              {recipe.create_time}分
            </span>
          )}
        </div>
        <div className="text-xs">
          食材{recipe.cooking.length}・調味料{recipe.season.length}・手順{stepCount}
        </div>
        <SharedAvatars users={recipe.shared_user} className="mt-1 self-start" />
      </CardContent>
    </Card>
  )
}

// 共有相手のアバターを重ねて並べる。共有が無いときも高さを確保し、
// 共有あり/なしでカードの大きさが変わらないようにする。
function SharedAvatars({ users, className }: { users: UserListItem[]; className?: string }) {
  const shown = users.slice(0, MAX_AVATARS)
  const overflow = users.length - shown.length

  return (
    <div className={cn('flex h-6 items-center', className)}>
      {shown.map((u, i) => (
        <Avatar key={u.id} className={cn('border-card size-6 border-2', i > 0 && '-ml-2')}>
          <AvatarImage src={u.avatar_url ?? undefined} alt={u.username} />
          <AvatarFallback className="text-[10px]">
            {u.username.charAt(0).toUpperCase()}
          </AvatarFallback>
        </Avatar>
      ))}
      {overflow > 0 && (
        <span className="bg-muted border-card -ml-2 flex size-6 items-center justify-center rounded-full border-2 text-[10px]">
          +{overflow}
        </span>
      )}
    </div>
  )
}
