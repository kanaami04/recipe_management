import { Link as LinkIcon, Loader2 } from 'lucide-react'
import { useRef, useState } from 'react'
import { toast } from 'sonner'

import { fetchOgp } from '@/shared/api/generated/sdk.gen'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'

import { RecipeSourceLink } from './RecipeSourceLink'

type Props = {
  url: string
  thumbnail: string
  onUrlChange: (value: string) => void
  onThumbnailChange: (value: string) => void
  // OGP から取れたタイトル。url モードでレシピ名の自動補完に使う(任意)。
  onTitleFetched?: (title: string) => void
}

// 参考にした外部レシピ(クラシル等)の URL 入力と、その OGP サムネイル表示。
// URL 入力を離れたタイミングでサーバ経由の OGP 取得を行い、画像 URL を親へ返す。
// サムネイルをタップすると元サイトを別タブで開く。
export function RecipeSourceInput({
  url,
  thumbnail,
  onUrlChange,
  onThumbnailChange,
  onTitleFetched,
}: Props) {
  const [loading, setLoading] = useState(false)
  // 直近に取得を試みた URL。同一 URL の再取得を避け、複数リクエストの順序逆転で
  // 古い結果が新しい URL のサムネ/タイトルを上書きするのを防ぐ(初期値は初回の url)。
  const lastFetchedUrl = useRef(url)

  // URL 入力欄を離れたら OGP を取得する。空なら取得済みサムネも消す。
  const loadThumbnail = async () => {
    const trimmed = url.trim()
    if (trimmed === '') {
      onThumbnailChange('')
      lastFetchedUrl.current = ''
      return
    }
    if (!/^https?:\/\//i.test(trimmed)) return
    if (trimmed === lastFetchedUrl.current) return // URL 未変更なら取り直さない
    lastFetchedUrl.current = trimmed
    setLoading(true)
    try {
      const { data } = await fetchOgp({ query: { url: trimmed }, throwOnError: true })
      // 取得中に別の URL が要求されていたら、この結果は破棄する(最後の要求だけ反映)。
      if (lastFetchedUrl.current !== trimmed) return
      onThumbnailChange(data.image)
      if (data.title !== '') onTitleFetched?.(data.title)
      if (data.image === '') {
        toast.info('このURLからサムネイル画像を取得できませんでした')
      }
    } catch (error) {
      console.error(error)
      toast.error('サムネイルの取得に失敗しました')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="grid gap-3">
      <Label className="flex items-center gap-1.5">
        <LinkIcon className="size-4" />
        参考レシピの URL
      </Label>
      <Input
        type="url"
        inputMode="url"
        placeholder="https://www.kurashiru.com/recipes/..."
        value={url}
        onChange={(e) => onUrlChange(e.target.value)}
        onBlur={loadThumbnail}
      />

      {loading && (
        <div className="text-muted-foreground flex items-center gap-2 text-sm">
          <Loader2 className="size-4 animate-spin" />
          サムネイルを取得中...
        </div>
      )}

      {!loading && thumbnail !== '' && (
        <RecipeSourceLink sourceUrl={url} thumbnail={thumbnail} label="開く" />
      )}
    </div>
  )
}
