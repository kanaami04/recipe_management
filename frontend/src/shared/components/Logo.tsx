import { cn } from '@/shared/lib/utils'

// cookience のフラスコ型マーク。ストロークは currentColor に追従し、
// 中の液体(ブランド色)だけ固定。文字色に載せればダーク/ライト両対応になる。
function LogoMark({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 100 100"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={cn('size-6', className)}
      aria-hidden="true"
    >
      <defs>
        <clipPath id="cookience-mark-interior">
          <path d="M45 22H55V42L70 76A6 6 0 0 1 64 82H36A6 6 0 0 1 30 76L45 42Z" />
        </clipPath>
      </defs>
      <rect
        x="24"
        y="64"
        width="52"
        height="30"
        className="fill-brand"
        clipPath="url(#cookience-mark-interior)"
      />
      <path d="M40 22H60" stroke="currentColor" strokeWidth="6" strokeLinecap="round" />
      <path
        d="M45 22V42L30 76A6 6 0 0 0 36 82H64A6 6 0 0 0 70 76L55 42V22"
        stroke="currentColor"
        strokeWidth="6"
        strokeLinejoin="round"
        strokeLinecap="round"
      />
      <path
        d="M48 49H54.8M48 55H57.5M48 61H60"
        stroke="currentColor"
        strokeWidth="3"
        strokeLinecap="round"
      />
    </svg>
  )
}

// cookience のワードマーク。全て小文字、語尾のピリオドだけブランド色。
// フォントは Space Grotesk(root.tsx で読み込み)。
function LogoWordmark({ className }: { className?: string }) {
  return (
    <span className={cn('font-brand leading-none font-semibold tracking-[-0.035em]', className)}>
      cookience<span className="text-brand">.</span>
    </span>
  )
}

// マーク + ワードマークの横組みロックアップ。
export function Logo({
  className,
  markClassName,
  wordmarkClassName,
}: {
  className?: string
  markClassName?: string
  wordmarkClassName?: string
}) {
  return (
    <span className={cn('inline-flex items-center gap-2', className)}>
      <LogoMark className={markClassName} />
      <LogoWordmark className={wordmarkClassName} />
    </span>
  )
}
