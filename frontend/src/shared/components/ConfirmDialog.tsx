import { cn } from '@/shared/lib/utils'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/shared/ui/alert-dialog'
import { buttonVariants } from '@/shared/ui/button'

type ConfirmDialogProps = {
  title: string
  description: string
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
  confirmLabel?: string
  cancelLabel?: string
  // 破壊的操作(削除など)では確定ボタンを destructive 配色にする。
  destructive?: boolean
}

// 確認系ダイアログ (ADR-0009 #3)。OK のみの通知は MessageAlertDialog を使う。
export function ConfirmDialog({
  title,
  description,
  open,
  onOpenChange,
  onConfirm,
  confirmLabel = 'OK',
  cancelLabel = 'キャンセル',
  destructive = false,
}: ConfirmDialogProps) {
  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription className="whitespace-pre-line">
            {description}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{cancelLabel}</AlertDialogCancel>
          <AlertDialogAction
            onClick={onConfirm}
            className={cn(destructive && buttonVariants({ variant: 'destructive' }))}
          >
            {confirmLabel}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
