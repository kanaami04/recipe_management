import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useRef, useState } from 'react'
import { toast } from 'sonner'

import {
  deleteAvatarMutation,
  getUserInfoQueryKey,
} from '@/shared/api/generated/@tanstack/react-query.gen'
import { confirmAvatar, createAvatarUploadUrl } from '@/shared/api/generated/sdk.gen'
import type {
  CreateAvatarUploadUrlRequest,
  UserInfoResponse,
} from '@/shared/api/generated/types.gen'
import { Avatar, AvatarFallback, AvatarImage } from '@/shared/ui/avatar'
import { Button } from '@/shared/ui/button'

import { AvatarCropDialog } from './AvatarCropDialog'

// アップロードを許可する画像形式(サーバの oneof と一致させる)と上限サイズ。
type ContentType = CreateAvatarUploadUrlRequest['content_type']
const ALLOWED_TYPES: ContentType[] = ['image/jpeg', 'image/png', 'image/webp']
const MAX_BYTES = 20 * 1024 * 1024 // 20MB

// プロフィール画像の変更・削除フォーム。
// アップロードは「署名付き URL 発行 → S3 へ直 PUT → 確定」の 3 段階。
export function AvatarForm({ user }: { user: UserInfoResponse }) {
  const queryClient = useQueryClient()
  const fileInputRef = useRef<HTMLInputElement>(null)
  // 選んだ画像を切り抜くための data URL。null のときはクロップダイアログを閉じる。
  const [cropSrc, setCropSrc] = useState<string | null>(null)

  const invalidate = () => queryClient.invalidateQueries({ queryKey: getUserInfoQueryKey() })

  const upload = useMutation({
    // 切り抜き済みの正方形 Blob(image/jpeg)を受け取ってアップロードする。
    mutationFn: async (blob: Blob) => {
      const contentType = blob.type as ContentType
      // 1) アップロード用の署名付き URL を得る(失敗時は throw して onError へ)
      const { data } = await createAvatarUploadUrl({
        body: { content_type: contentType },
        throwOnError: true,
      })
      // 2) 署名付き URL へ画像本体を直 PUT する(認証ヘッダなしの素の fetch)。
      //    Content-Type は署名時と同じ値にする。
      const res = await fetch(data.upload_url, {
        method: 'PUT',
        headers: { 'Content-Type': contentType },
        body: blob,
      })
      if (!res.ok) throw new Error('アップロードに失敗しました')
      // 3) アップロード済みの key をプロフィール画像として確定する
      await confirmAvatar({ body: { key: data.key }, throwOnError: true })
    },
    onSuccess: () => {
      invalidate()
      setCropSrc(null)
      toast.success('プロフィール画像を変更しました')
    },
    onError: () => toast.error('プロフィール画像の変更に失敗しました'),
  })

  const removeAvatar = useMutation({
    ...deleteAvatarMutation(),
    onSuccess: () => {
      invalidate()
      toast.success('プロフィール画像を削除しました')
    },
    onError: () => toast.error('プロフィール画像の削除に失敗しました'),
  })

  const onFileSelected = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    // 同じファイルを選び直しても change が発火するよう value をリセットする。
    e.target.value = ''
    if (!file) return
    if (!ALLOWED_TYPES.includes(file.type as ContentType)) {
      toast.error('対応している形式は JPEG / PNG / WebP です')
      return
    }
    if (file.size > MAX_BYTES) {
      toast.error('画像は 20MB 以下にしてください')
      return
    }
    // 画像本体はまだ送らず、data URL に変換してクロップダイアログを開く。
    const reader = new FileReader()
    reader.addEventListener('load', () => setCropSrc(reader.result as string))
    reader.addEventListener('error', () => {
      console.error(reader.error)
      toast.error('画像の読み込みに失敗しました')
    })
    reader.readAsDataURL(file)
  }

  const busy = upload.isPending || removeAvatar.isPending

  return (
    <>
      <div className="flex items-center gap-4">
        <Avatar className="size-16">
          <AvatarImage src={user.avatar_url ?? undefined} alt="プロフィール画像" />
          <AvatarFallback className="text-lg">
            {user.username.charAt(0).toUpperCase()}
          </AvatarFallback>
        </Avatar>
        <div className="flex flex-col gap-2">
          <input
            ref={fileInputRef}
            type="file"
            accept="image/jpeg,image/png,image/webp"
            className="hidden"
            onChange={onFileSelected}
          />
          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              disabled={busy}
              onClick={() => fileInputRef.current?.click()}
            >
              画像を変更
            </Button>
            {user.avatar_url && (
              <Button
                type="button"
                variant="ghost"
                disabled={busy}
                onClick={() => removeAvatar.mutate({})}
              >
                削除
              </Button>
            )}
          </div>
          <p className="text-muted-foreground text-sm">JPEG / PNG / WebP、20MB まで。</p>
        </div>
      </div>
      <AvatarCropDialog
        key={cropSrc ?? 'closed'}
        imageSrc={cropSrc}
        busy={upload.isPending}
        onCancel={() => setCropSrc(null)}
        onCropped={(blob) => upload.mutate(blob)}
      />
    </>
  )
}
