import { expect, type Page, test } from '@playwright/test'

// アバター画像を選ぶとクロップ編集ダイアログが開き、切り抜いてアップロードできることを確認する。
async function mockApi(page: Page) {
  await page.route('**/api/token/', (route) =>
    route.fulfill({ status: 200, json: { access: 'fake-access' } }),
  )
  await page.route('**/api/token/refresh/', (route) =>
    route.fulfill({ status: 200, json: { access: 'fake-access' } }),
  )
  await page.route('**/api/user_info/', (route) =>
    route.fulfill({
      status: 200,
      json: { id: 'u-taro', username: 'taro', email: 'taro@example.com' },
    }),
  )
  // アバターアップロードの 3 段階(URL 発行=POST → S3 直 PUT → 確定=PUT)をモックする。
  // 発行と確定は同じ /api/user_info/avatar/ をメソッドで分岐する。
  await page.route('**/api/user_info/avatar/', (route) => {
    if (route.request().method() === 'POST') {
      return route.fulfill({
        status: 200,
        json: { key: 'avatars/u-taro.jpg', upload_url: 'https://s3.example/put' },
      })
    }
    return route.fulfill({ status: 200, json: {} }) // PUT(確定)
  })
  await page.route('https://s3.example/put', (route) => route.fulfill({ status: 200, body: '' }))
}

// 横長(2:1)の PNG を data URL で用意する。切り抜き前の比率崩れを再現しやすい形。
const WIDE_PNG =
  'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAEAAAAAgCAYAAACinX6EAAAAKklEQVR4nO3BAQ0AAADCoPdPbQ8HFAAAAAAAAAAAAAAAAAAAAAAAAADwbQ8AAAHY7pQ0AAAAAElFTkSuQmCC'

test('アバター画像を選ぶとクロップ編集ダイアログが出る', async ({ page }) => {
  await mockApi(page)
  await page.goto('/')
  await page.fill('#email', 'taro@example.com')
  await page.fill('#password', 'password123')
  await page.getByRole('button', { name: 'ログイン' }).click()
  await expect(page).toHaveURL(/\/top$/)

  await page.goto('/top/account')

  // 横長画像を file input に流し込む。
  const buffer = Buffer.from(WIDE_PNG.split(',')[1], 'base64')
  await page.setInputFiles('input[type="file"]', {
    name: 'wide.png',
    mimeType: 'image/png',
    buffer,
  })

  await expect(page.getByRole('heading', { name: 'アイコンを切り抜く' })).toBeVisible()
  const confirm = page.getByRole('button', { name: '決定' })
  // クロッパの初期描画で croppedAreaPixels が入り、決定が押せるようになるのを待つ。
  await expect(confirm).toBeEnabled()

  // 決定 → 切り抜き(canvas)→ 3 段アップロード → 成功トースト → ダイアログが閉じる。
  await confirm.click()
  await expect(page.getByText('プロフィール画像を変更しました')).toBeVisible()
  await expect(page.getByRole('heading', { name: 'アイコンを切り抜く' })).toBeHidden()
})
