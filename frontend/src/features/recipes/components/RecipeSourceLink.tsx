import { ExternalLink } from 'lucide-react'

// 参考にした外部レシピへのリンク。サムネイル画像があれば画像パネル、無ければ
// テキストリンクで表示する。いずれもタップで元サイトを別タブで開く。
// フォームのプレビュー・詳細画面の両方で使う共通コンポーネント。
export function RecipeSourceLink({
  sourceUrl,
  thumbnail,
  label,
}: {
  sourceUrl: string
  thumbnail: string
  label: string
}) {
  if (thumbnail === '') {
    return (
      <a
        href={sourceUrl}
        target="_blank"
        rel="noreferrer"
        className="text-primary inline-flex w-fit items-center gap-1 text-sm underline underline-offset-2"
      >
        <ExternalLink className="size-3.5" />
        {label}
      </a>
    )
  }

  return (
    <a
      href={sourceUrl}
      target="_blank"
      rel="noreferrer"
      className="group ring-border hover:ring-primary/60 relative block overflow-hidden rounded-lg ring-1 transition"
    >
      <img
        src={thumbnail}
        alt="参考レシピのサムネイル"
        className="aspect-video w-full object-cover"
        loading="lazy"
      />
      {/* タップで別タブ遷移することを示す控えめなヒント。 */}
      <span className="bg-background/85 text-foreground absolute right-2 bottom-2 flex items-center gap-1 rounded-full px-2.5 py-1 text-xs shadow-sm">
        <ExternalLink className="size-3.5" />
        {label}
      </span>
    </a>
  )
}
