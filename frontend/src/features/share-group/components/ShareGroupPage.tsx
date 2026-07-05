import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import { Copy, RefreshCw, Trash2 } from 'lucide-react'
import { useState } from 'react'
import { toast } from 'sonner'

import {
  createShareGroupMutation,
  getShareGroupOptions,
  getShareGroupQueryKey,
  getShoppingListQueryKey,
  joinShareGroupMutation,
  leaveShareGroupMutation,
  listRecipesQueryKey,
  regenerateInviteCodeMutation,
  removeShareGroupMemberMutation,
} from '@/shared/api/generated/@tanstack/react-query.gen'
import type { ShareGroupResponse, UserListItem } from '@/shared/api/generated/types.gen'
import { ConfirmDialog } from '@/shared/components/ConfirmDialog'
import { Avatar, AvatarFallback, AvatarImage } from '@/shared/ui/avatar'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Separator } from '@/shared/ui/separator'
import { SidebarTrigger } from '@/shared/ui/sidebar'

// 共有グループ画面。未所属ならグループ作成/参加のオンボーディングを、所属済みなら
// メンバー・招待コード・脱退の管理を表示する。グループのメンバーはレシピ・買い物リストを
// 全員で自動共有する。
export function ShareGroupPage() {
  const queryClient = useQueryClient()
  // 未所属は 404。404 は即座にオンボーディングへ倒し、それ以外(一時的な失敗)は数回リトライする。
  const {
    data: group,
    isPending,
    error,
  } = useQuery({
    ...getShareGroupOptions(),
    retry: (count, err) => (err as AxiosError).response?.status !== 404 && count < 2,
  })
  const notFound = (error as AxiosError | null)?.response?.status === 404

  // グループの有無で可視範囲が変わるため、レシピ・買い物リストのキャッシュも無効化する。
  const invalidateShared = () => {
    queryClient.invalidateQueries({ queryKey: listRecipesQueryKey() })
    queryClient.invalidateQueries({ queryKey: getShoppingListQueryKey() })
  }
  // 作成/参加/再発行の結果(グループ本体)を即反映する。invalidate 頼みだと 404 error 状態のまま
  // オンボーディングが一瞬再表示されるため、返ってきたグループを直接キャッシュへ入れる。
  const syncGroup = (g: ShareGroupResponse) => {
    queryClient.setQueryData(getShareGroupQueryKey(), g)
    invalidateShared()
  }
  // メンバー削除(グループは残る)の後にグループを再取得する。
  const refetchGroup = () => {
    queryClient.invalidateQueries({ queryKey: getShareGroupQueryKey() })
    invalidateShared()
  }
  // 脱退/解散(グループが無くなる)後は reset で初期状態へ戻して再取得させる。
  // invalidate だと 404 になっても直前のグループ data が残り、オンボーディングへ切り替わらない。
  // reset は data を捨てたうえでアクティブな observer を再取得させるため、404 → オンボーディングに移る。
  const clearGroup = () => {
    queryClient.resetQueries({ queryKey: getShareGroupQueryKey() })
    invalidateShared()
  }

  return (
    <>
      <header className="flex h-(--header-height) shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear">
        <div className="flex w-full items-center gap-2 px-3 sm:px-4 lg:px-6">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mx-1 data-[orientation=vertical]:h-4" />
          <h1 className="text-base font-medium">共有グループ</h1>
        </div>
      </header>

      <div className="mx-auto flex w-full max-w-2xl flex-col gap-4 p-3 sm:p-4">
        {isPending ? (
          <p className="text-muted-foreground py-8 text-center">読み込み中...</p>
        ) : group ? (
          <ShareGroupDetail
            group={group}
            onGroupUpdated={syncGroup}
            onMemberRemoved={refetchGroup}
            onLeft={clearGroup}
          />
        ) : notFound ? (
          <ShareGroupOnboarding onJoined={syncGroup} />
        ) : (
          <p className="text-destructive py-8 text-center">共有グループの取得に失敗しました</p>
        )}
      </div>
    </>
  )
}

// 未所属のときのオンボーディング。グループ作成 or 招待コードで参加。
function ShareGroupOnboarding({ onJoined }: { onJoined: (g: ShareGroupResponse) => void }) {
  const [name, setName] = useState('')
  const [code, setCode] = useState('')

  const create = useMutation({
    ...createShareGroupMutation(),
    onSuccess: (data) => {
      onJoined(data)
      toast.success('共有グループを作成しました')
    },
    onError: () => toast.error('グループの作成に失敗しました'),
  })

  const join = useMutation({
    ...joinShareGroupMutation(),
    onSuccess: (data) => {
      onJoined(data)
      toast.success('共有グループに参加しました')
    },
    onError: (err) => {
      const message =
        err.response?.status === 409
          ? '既に別のグループに参加しています'
          : err.response?.status === 400
            ? '招待コードが無効か期限切れです'
            : 'グループへの参加に失敗しました'
      toast.error(message)
    },
  })

  const submitJoin = () => {
    if (code.trim() === '' || join.isPending) return
    join.mutate({ body: { invite_code: code.trim() } })
  }

  return (
    <div className="flex flex-col gap-6">
      <p className="text-muted-foreground text-sm">
        共有グループを作ると、レシピや買い物リストをメンバー全員で共有できます。
        まずはグループを作るか、招待コードで既存のグループに参加してください。
      </p>

      <div className="flex flex-col gap-2 rounded-md border p-4">
        <h2 className="font-medium">グループを作成</h2>
        <div className="flex gap-2">
          <Input
            placeholder="グループ名(例: 我が家)"
            value={name}
            maxLength={50}
            onChange={(e) => setName(e.target.value)}
          />
          <Button
            onClick={() => create.mutate({ body: { name: name.trim() } })}
            disabled={create.isPending}
          >
            作成
          </Button>
        </div>
      </div>

      <div className="flex flex-col gap-2 rounded-md border p-4">
        <h2 className="font-medium">招待コードで参加</h2>
        <div className="flex gap-2">
          <Input
            placeholder="招待コード"
            value={code}
            onChange={(e) => setCode(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') submitJoin()
            }}
          />
          <Button
            variant="outline"
            onClick={submitJoin}
            disabled={join.isPending || code.trim() === ''}
          >
            参加
          </Button>
        </div>
      </div>
    </div>
  )
}

// 所属済みのときの管理画面。
function ShareGroupDetail({
  group,
  onGroupUpdated,
  onMemberRemoved,
  onLeft,
}: {
  group: ShareGroupResponse
  onGroupUpdated: (g: ShareGroupResponse) => void
  onMemberRemoved: () => void
  onLeft: () => void
}) {
  const [removeTarget, setRemoveTarget] = useState<UserListItem | null>(null)
  const [leaveOpen, setLeaveOpen] = useState(false)

  const regenerate = useMutation({
    ...regenerateInviteCodeMutation(),
    onSuccess: (data) => {
      onGroupUpdated(data)
      toast.success('招待コードを再発行しました')
    },
    onError: () => toast.error('招待コードの再発行に失敗しました'),
  })

  const removeMember = useMutation({
    ...removeShareGroupMemberMutation(),
    onSuccess: () => {
      onMemberRemoved()
      setRemoveTarget(null)
      toast.success('メンバーを外しました')
    },
    onError: () => toast.error('メンバーを外せませんでした'),
  })

  const leave = useMutation({
    ...leaveShareGroupMutation(),
    onSuccess: () => {
      onLeft()
      setLeaveOpen(false)
      toast.success(group.is_owner ? 'グループを解散しました' : 'グループを抜けました')
    },
    onError: () => toast.error(group.is_owner ? '解散に失敗しました' : '退出に失敗しました'),
  })

  const copyCode = async () => {
    try {
      await navigator.clipboard.writeText(group.invite_code)
      toast.success('招待コードをコピーしました')
    } catch {
      toast.error('コピーできませんでした')
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-lg font-medium">{group.name}</h2>
        <p className="text-muted-foreground text-sm">
          メンバーはレシピ・買い物リストを全員で共有します。
        </p>
      </div>

      {/* メンバー一覧 */}
      <div className="flex flex-col gap-2">
        <h3 className="text-sm font-medium">メンバー({group.members.length})</h3>
        <ul className="divide-border divide-y rounded-md border">
          {group.members.map((member) => (
            <li key={member.id} className="flex items-center gap-3 p-3">
              <Avatar className="size-8">
                <AvatarImage src={member.avatar_url ?? undefined} alt={member.username} />
                <AvatarFallback>{member.username.charAt(0).toUpperCase()}</AvatarFallback>
              </Avatar>
              <span className="flex-1 truncate">{member.username}</span>
              {member.id === group.owner.id ? (
                <span className="text-muted-foreground text-xs">オーナー</span>
              ) : (
                group.is_owner && (
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label="メンバーを外す"
                    onClick={() => setRemoveTarget(member)}
                  >
                    <Trash2 />
                  </Button>
                )
              )}
            </li>
          ))}
        </ul>
      </div>

      {/* 招待コード */}
      <div className="flex flex-col gap-2">
        <h3 className="text-sm font-medium">招待コード</h3>
        <div className="flex items-center gap-2">
          <code className="bg-muted rounded px-3 py-2 font-mono text-base tracking-widest">
            {group.invite_code}
          </code>
          <Button variant="outline" size="icon" aria-label="コードをコピー" onClick={copyCode}>
            <Copy />
          </Button>
          {group.is_owner && (
            <Button
              variant="outline"
              size="icon"
              aria-label="コードを再発行"
              disabled={regenerate.isPending}
              onClick={() => regenerate.mutate({})}
            >
              <RefreshCw />
            </Button>
          )}
        </div>
        <p className="text-muted-foreground text-xs">
          このコードを渡すと相手が参加できます(有効期限: {group.invite_code_expires_at})。
          {group.is_owner && ' 再発行すると古いコードは使えなくなります。'}
        </p>
      </div>

      {/* 脱退 / 解散 */}
      <div className="flex justify-end border-t pt-4">
        <Button variant="outline" onClick={() => setLeaveOpen(true)} disabled={leave.isPending}>
          {group.is_owner ? 'グループを解散' : 'グループを抜ける'}
        </Button>
      </div>

      <ConfirmDialog
        title={removeTarget ? `${removeTarget.username} を外しますか？` : ''}
        description="このメンバーは共有物が見られなくなります。"
        open={removeTarget !== null}
        onOpenChange={(open) => !open && setRemoveTarget(null)}
        onConfirm={() =>
          removeTarget && removeMember.mutate({ path: { user_id: removeTarget.id } })
        }
        confirmLabel="外す"
        destructive
      />

      <ConfirmDialog
        title={group.is_owner ? 'グループを解散しますか？' : 'グループを抜けますか？'}
        description={
          group.is_owner
            ? 'グループが解散され、メンバー全員が共有物を見られなくなります。'
            : '抜けると、このグループの共有物が見られなくなります。'
        }
        open={leaveOpen}
        onOpenChange={setLeaveOpen}
        onConfirm={() => leave.mutate({})}
        confirmLabel={group.is_owner ? '解散' : '抜ける'}
        destructive
      />
    </div>
  )
}
