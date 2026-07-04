import { useNavigate } from 'react-router-dom'

import { useSidebar } from '@/shared/ui/sidebar'

// モバイルでは、遷移と同時に開いていたサイドバー(Sheet)を閉じる。
// Sheet(Radix Dialog)の通常のクローズアニメーション(300ms)の間、開いた時に
// 掛けた body の pointer-events:none が Radix 側でまだ解除されておらず、
// 遷移先の画面をタップしても反応しない一瞬が生じる。アニメーションの完了を待たず
// 即座に解除し、遷移直後から操作できるようにする。
export function useNavigateAndClose() {
  const navigate = useNavigate()
  const { isMobile, setOpenMobile } = useSidebar()

  return (to: string) => {
    navigate(to)
    if (isMobile) {
      setOpenMobile(false)
      requestAnimationFrame(() => {
        document.body.style.pointerEvents = ''
      })
    }
  }
}
