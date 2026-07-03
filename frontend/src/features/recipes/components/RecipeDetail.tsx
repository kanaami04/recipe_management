import { Carrot, Clock, Droplet, ListOrdered, Users } from 'lucide-react'
import type { ReactNode } from 'react'

import { splitSteps } from '@/features/recipes/steps'
import { formatAmount } from '@/features/recipes/units'
import type { RecipeResponse } from '@/shared/api/generated/types.gen'
import { Badge } from '@/shared/ui/badge'

// 詳細ダイアログの本文。画像を持たないテキスト中心のレシピを、
// メタ情報チップ・右揃えの分量・番号付き手順で読みやすく組む(レシピ投稿アプリの定番構成)。
export function RecipeDetail({ recipe }: { recipe: RecipeResponse }) {
  const steps = splitSteps(recipe.procedure).filter((s) => s.trim() !== '')

  return (
    <div className="flex flex-col gap-6">
      {/* クイックファクト。人数・調理時間・材料数をチップで一望させる。 */}
      <div className="flex flex-wrap gap-2">
        <MetaChip icon={<Users className="size-3.5" />}>{recipe.create_for}人前</MetaChip>
        {recipe.create_time != null && (
          <MetaChip icon={<Clock className="size-3.5" />}>{recipe.create_time}分</MetaChip>
        )}
        <MetaChip icon={<Carrot className="size-3.5" />}>材料{recipe.cooking.length}品</MetaChip>
      </div>

      {recipe.label.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {recipe.label.map((label) => (
            <Badge key={label.name} variant="secondary">
              {label.name}
            </Badge>
          ))}
        </div>
      )}

      {recipe.cooking.length > 0 && (
        <Section
          icon={<Carrot className="size-4" />}
          title="材料"
          note={`${recipe.create_for}人前`}
        >
          <AmountList
            items={recipe.cooking.map((c) => ({
              name: c.ingredients.name,
              amount: formatAmount(c.quantity, c.unit),
            }))}
          />
        </Section>
      )}

      {recipe.season.length > 0 && (
        <Section icon={<Droplet className="size-4" />} title="調味料">
          <AmountList
            items={recipe.season.map((s) => ({
              name: s.seasoning.name,
              amount: formatAmount(s.quantity, s.unit),
            }))}
          />
        </Section>
      )}

      {steps.length > 0 && (
        <Section icon={<ListOrdered className="size-4" />} title="作り方">
          <ol className="flex flex-col gap-3">
            {steps.map((step, index) => (
              <li key={index} className="flex gap-3">
                <span className="bg-primary text-primary-foreground mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-full text-xs font-medium tabular-nums">
                  {index + 1}
                </span>
                <p className="leading-relaxed break-words whitespace-pre-wrap">{step}</p>
              </li>
            ))}
          </ol>
        </Section>
      )}

      {recipe.shared_user.length > 0 && (
        <Section title="共有ユーザー">
          <div className="flex flex-wrap gap-1.5">
            {recipe.shared_user.map((user) => (
              <Badge key={user.id} variant="outline">
                {user.username}
              </Badge>
            ))}
          </div>
        </Section>
      )}

      <p className="text-muted-foreground text-xs">更新: {recipe.updated_at}</p>
    </div>
  )
}

// メタ情報チップ。アイコン + テキストの pill 表示。
function MetaChip({ icon, children }: { icon: ReactNode; children: ReactNode }) {
  return (
    <span className="bg-muted text-muted-foreground flex items-center gap-1 rounded-full px-3 py-1 text-sm">
      {icon}
      {children}
    </span>
  )
}

// 見出し付きセクション。任意の補足(人数など)を見出し内に控えめに添える。
function Section({
  icon,
  title,
  note,
  children,
}: {
  icon?: ReactNode
  title: string
  note?: string
  children: ReactNode
}) {
  return (
    <section className="flex flex-col gap-2">
      <h3 className="flex items-center gap-1.5 text-sm font-semibold">
        {icon}
        {title}
        {note && <span className="text-muted-foreground font-normal">({note})</span>}
      </h3>
      {children}
    </section>
  )
}

// 名前=左・分量=右揃え、行ごとに区切り線で並べる材料/調味料リスト。
function AmountList({ items }: { items: { name: string; amount: string }[] }) {
  return (
    <ul className="divide-border divide-y">
      {items.map((item) => (
        <li key={item.name} className="flex items-baseline justify-between gap-4 py-2 text-sm">
          <span>{item.name}</span>
          <span className="text-muted-foreground shrink-0 tabular-nums">{item.amount}</span>
        </li>
      ))}
    </ul>
  )
}
