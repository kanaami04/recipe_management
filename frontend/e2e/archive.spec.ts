import { expect, type Page, test } from '@playwright/test'

const mkRecipe = (id: string, title: string, archived: boolean) => ({
  id,
  created_at: '2026-06-15 09:30',
  updated_at: '2026-06-15 09:30',
  cooking: [{ ingredients: { name: '玉ねぎ' }, quantity: 1, unit: '個' }],
  season: [],
  procedure: '煮る',
  owner: { id: 'u-taro', username: 'taro' },
  shared_user: [],
  title,
  create_time: 30,
  create_for: 2,
  archive_flg: archived,
  label: [],
})

async function mockApi(page: Page) {
  const recipes = [mkRecipe('r1', 'カレー', false), mkRecipe('r2', 'サラダ', false)]
  await page.route('**/api/token/', (r) => r.fulfill({ status: 200, json: { access: 'a' } }))
  await page.route('**/api/token/refresh/', (r) =>
    r.fulfill({ status: 200, json: { access: 'a' } }),
  )
  await page.route('**/api/user_info/', (r) =>
    r.fulfill({ status: 200, json: { id: 'u-taro', username: 'taro', email: 'taro@example.com' } }),
  )
  await page.route('**/api/label/', (r) => r.fulfill({ status: 200, json: [] }))
  await page.route('**/api/users/', (r) => r.fulfill({ status: 200, json: [] }))
  // PUT は archive_flg を反映して以降の一覧へ反映する。
  await page.route('**/api/recipes/*/', (r) => {
    const id = r
      .request()
      .url()
      .match(/\/api\/recipes\/([^/]+)\//)?.[1]
    const body = r.request().postDataJSON()
    const idx = recipes.findIndex((x) => x.id === id)
    if (r.request().method() === 'PUT' && idx >= 0) {
      recipes[idx] = { ...recipes[idx], archive_flg: body.archive_flg }
      return r.fulfill({ status: 200, json: recipes[idx] })
    }
    return r.fallback()
  })
  await page.route('**/api/recipes/', (r) => r.fulfill({ status: 200, json: recipes }))
}

async function login(page: Page) {
  await page.goto('/')
  await page.fill('#email', 'taro@example.com')
  await page.fill('#password', 'password123')
  await page.getByRole('button', { name: 'ログイン' }).click()
  await page.waitForURL(/\/top$/)
}

test('レシピをアーカイブするとメインから消え、アーカイブ一覧に現れる', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await expect(page.getByText('カレー')).toBeVisible()

  // Act: カレーの詳細 → ⋮ → アーカイブする
  await page.getByText('カレー').click()
  await page.getByRole('button', { name: '操作メニュー' }).click()
  await page.getByRole('menuitem', { name: 'アーカイブする' }).click()

  // Assert: 成功トーストが出て、メイン一覧からカレーが消える
  await expect(page.getByText('レシピをアーカイブしました')).toBeVisible()
  await expect(page.getByText('カレー')).toHaveCount(0)
  await expect(page.getByText('サラダ')).toBeVisible()

  // Act: サイドバーのアーカイブへ
  await page.getByRole('button', { name: 'アーカイブ', exact: true }).click()

  // Assert: アーカイブ一覧にカレーが出る
  await expect(page).toHaveURL(/\/top\/archive$/)
  await expect(page.getByText('カレー')).toBeVisible()
})

test('スマホでサイドバーから遷移するとサイドバーが閉じる', async ({ page }) => {
  // Arrange: モバイル幅(サイドバーが Sheet)でログイン
  await page.setViewportSize({ width: 375, height: 800 })
  await mockApi(page)
  await login(page)

  // Act: サイドバーを開き、アーカイブへ遷移する
  await page.getByRole('button', { name: 'Toggle Sidebar' }).click()
  await page.getByRole('button', { name: 'アーカイブ', exact: true }).click()

  // Assert: 遷移し、サイドバー(Sheet=dialog)が閉じる
  await expect(page).toHaveURL(/\/top\/archive$/)
  await expect(page.getByRole('dialog')).toHaveCount(0)
})

test('アーカイブ済みを解除するとメインに戻る', async ({ page }) => {
  // Arrange: 先にカレーをアーカイブしてアーカイブ一覧を開く
  await mockApi(page)
  await login(page)
  await page.getByText('カレー').click()
  await page.getByRole('button', { name: '操作メニュー' }).click()
  await page.getByRole('menuitem', { name: 'アーカイブする' }).click()
  await expect(page.getByText('レシピをアーカイブしました')).toBeVisible()
  await page.getByRole('button', { name: 'アーカイブ', exact: true }).click()
  await expect(page.getByText('カレー')).toBeVisible()

  // Act: アーカイブ一覧でカレー → ⋮ → アーカイブを解除
  await page.getByText('カレー').click()
  await page.getByRole('button', { name: '操作メニュー' }).click()
  await page.getByRole('menuitem', { name: 'アーカイブを解除' }).click()

  // Assert: 解除トーストが出て、アーカイブ一覧から消える
  await expect(page.getByText('アーカイブを解除しました')).toBeVisible()
  await expect(page.getByText('カレー')).toHaveCount(0)
})
