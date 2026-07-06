import type { Area } from 'react-easy-crop'

// アイコン編集画面で選んだ範囲(croppedAreaPixels)を、正方形の画像 Blob に切り抜く。
// 出力は一辺 OUTPUT_SIZE の JPEG。元画像はローカルの data URL なので canvas は汚染されない。
const OUTPUT_SIZE = 512

export async function getCroppedBlob(imageSrc: string, area: Area): Promise<Blob> {
  const image = await loadImage(imageSrc)
  const canvas = document.createElement('canvas')
  canvas.width = OUTPUT_SIZE
  canvas.height = OUTPUT_SIZE
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error('canvas コンテキストを取得できませんでした')

  // 選択範囲(area)を canvas 全体へ拡縮して描き込むと、正方形に切り抜かれる。
  ctx.drawImage(image, area.x, area.y, area.width, area.height, 0, 0, OUTPUT_SIZE, OUTPUT_SIZE)

  return new Promise((resolve, reject) => {
    canvas.toBlob(
      (blob) => (blob ? resolve(blob) : reject(new Error('画像の切り抜きに失敗しました'))),
      'image/jpeg',
      0.9,
    )
  })
}

function loadImage(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image()
    image.addEventListener('load', () => resolve(image))
    image.addEventListener('error', () => reject(new Error('画像の読み込みに失敗しました')))
    image.src = src
  })
}
