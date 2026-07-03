import { expect, type Page, test } from '@playwright/test'

// 並び替え検証用に 2 件のレシピを持つモック。
const mkRecipe = (id: string, title: string) => ({
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
  archive_flg: false,
  label: [],
})

async function mockApi(page: Page, onReorder: (ids: string[]) => void) {
  const recipes = [mkRecipe('r1', 'カレー'), mkRecipe('r2', 'サラダ')]
  await page.route('**/api/token/', (r) => r.fulfill({ status: 200, json: { access: 'a' } }))
  await page.route('**/api/token/refresh/', (r) =>
    r.fulfill({ status: 200, json: { access: 'a' } }),
  )
  await page.route('**/api/user_info/', (r) =>
    r.fulfill({ status: 200, json: { id: 'u-taro', username: 'taro', email: 'taro@example.com' } }),
  )
  await page.route('**/api/label/', (r) => r.fulfill({ status: 200, json: [] }))
  await page.route('**/api/users/', (r) => r.fulfill({ status: 200, json: [] }))
  await page.route('**/api/recipes/reorder/', (r) => {
    const ids: string[] = r.request().postDataJSON().recipe_ids
    onReorder(ids)
    // 実バックエンド同様、以降の一覧はこの並びで返す。
    recipes.sort((a, b) => ids.indexOf(a.id) - ids.indexOf(b.id))
    return r.fulfill({ status: 204, body: '' })
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

// カード名の表示順を返す。
async function cardOrder(page: Page): Promise<string[]> {
  return page.locator('[data-slot="card-title"]').allInnerTexts()
}

test('ハンドルをドラッグしてレシピを並び替えると順序が保存される', async ({ page }) => {
  // Arrange
  let sentOrder: string[] | null = null
  await mockApi(page, (ids) => {
    sentOrder = ids
  })
  await login(page)
  await expect(page.getByText('カレー')).toBeVisible()
  expect(await cardOrder(page)).toEqual(['カレー', 'サラダ'])

  // Act: 1 枚目(カレー)のカードを 2 枚目(サラダ)の位置までドラッグする
  const from = await page.getByText('カレー').boundingBox()
  const target = page.getByText('サラダ')
  const to = await target.boundingBox()
  if (!from || !to) throw new Error('bounding box not found')
  await page.mouse.move(from.x + from.width / 2, from.y + from.height / 2)
  await page.mouse.down()
  // 8px の閾値を超えてドラッグ開始 → ターゲットへ移動 → ドロップ
  await page.mouse.move(from.x + 20, from.y + from.height / 2, { steps: 5 })
  await page.mouse.move(to.x + to.width / 2, to.y + to.height / 2, { steps: 10 })
  await page.mouse.up()

  // Assert: 表示順が入れ替わり、その並びで reorder API が呼ばれる
  await expect(async () => {
    expect(await cardOrder(page)).toEqual(['サラダ', 'カレー'])
  }).toPass()
  expect(sentOrder).toEqual(['r2', 'r1'])
})
