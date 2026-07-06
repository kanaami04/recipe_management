import { useCallback, useState } from 'react'
import Cropper, { type Area, type Point } from 'react-easy-crop'
import { toast } from 'sonner'

import { Button } from '@/shared/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/shared/ui/dialog'

import { getCroppedBlob } from '../lib/cropImage'

// 画像を選んだあとに開く、アイコンの切り抜き編集ダイアログ。
// 円形の枠に合わせてドラッグ(位置)とスライダー(拡大)で調整し、正方形の Blob を返す。
// imageSrc が null のときは閉じている。呼び出し側は選び直しごとに key で作り直す想定。
export function AvatarCropDialog({
  imageSrc,
  busy,
  onCancel,
  onCropped,
}: {
  imageSrc: string | null
  busy: boolean
  onCancel: () => void
  onCropped: (blob: Blob) => void
}) {
  const [crop, setCrop] = useState<Point>({ x: 0, y: 0 })
  const [zoom, setZoom] = useState(1)
  const [croppedAreaPixels, setCroppedAreaPixels] = useState<Area | null>(null)
  const [cropping, setCropping] = useState(false)

  const onCropComplete = useCallback((_: Area, pixels: Area) => {
    setCroppedAreaPixels(pixels)
  }, [])

  const disabled = busy || cropping

  const handleConfirm = async () => {
    if (!imageSrc || !croppedAreaPixels) return
    setCropping(true)
    try {
      const blob = await getCroppedBlob(imageSrc, croppedAreaPixels)
      onCropped(blob)
    } catch (error) {
      console.error(error)
      toast.error('画像の切り抜きに失敗しました')
    } finally {
      setCropping(false)
    }
  }

  return (
    <Dialog
      open={imageSrc !== null}
      onOpenChange={(open) => {
        if (!open && !disabled) onCancel()
      }}
    >
      <DialogContent showCloseButton={false}>
        <DialogHeader>
          <DialogTitle>アイコンを切り抜く</DialogTitle>
          <DialogDescription>ドラッグで位置、スライダーで拡大を調整できます。</DialogDescription>
        </DialogHeader>
        <div className="bg-muted relative h-64 w-full overflow-hidden rounded-md sm:h-72">
          {imageSrc && (
            <Cropper
              image={imageSrc}
              crop={crop}
              zoom={zoom}
              aspect={1}
              cropShape="round"
              showGrid={false}
              onCropChange={setCrop}
              onZoomChange={setZoom}
              onCropComplete={onCropComplete}
            />
          )}
        </div>
        <div className="flex items-center gap-3">
          <span className="text-muted-foreground shrink-0 text-sm">拡大</span>
          <input
            type="range"
            min={1}
            max={3}
            step={0.01}
            value={zoom}
            onChange={(e) => setZoom(Number(e.target.value))}
            className="accent-primary w-full"
            aria-label="拡大"
            disabled={disabled}
          />
        </div>
        <DialogFooter>
          <Button type="button" variant="ghost" onClick={onCancel} disabled={disabled}>
            キャンセル
          </Button>
          <Button
            type="button"
            onClick={handleConfirm}
            disabled={disabled || croppedAreaPixels === null}
          >
            {disabled ? '保存中…' : '決定'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
