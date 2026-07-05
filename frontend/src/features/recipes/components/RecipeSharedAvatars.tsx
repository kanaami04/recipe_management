import type { UserListItem } from '@/shared/api/generated/types.gen'
import { cn } from '@/shared/lib/utils'
import { Avatar, AvatarFallback, AvatarImage } from '@/shared/ui/avatar'

// 一覧カードに並べる共有相手アバターの最大数。超過分は「+N」でまとめる
// (カード幅・高さを共有人数に依らず一定に保つため)。
const MAX_AVATARS = 3

// 共有相手のアバターを重ねて並べる。共有が無いときも高さ(h-6)を確保し、
// 共有あり/なしでカードの大きさが変わらないようにする。
export function RecipeSharedAvatars({
  users,
  className,
}: {
  users: UserListItem[]
  className?: string
}) {
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
