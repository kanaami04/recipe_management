import { type Page, test } from '@playwright/test'

// 一覧グリッドが分かるよう複数レシピをモックする(API 契約 RecipeResponse の形)。
const titles = ['カレー', '肉じゃが', '味噌汁', '唐揚げ', '親子丼', 'オムライス']
const recipes = titles.map((title, i) => ({
  id: `018f1a2b-3c4d-7e5f-8a9b-00000000000${i + 1}`,
  created_at: '2026-06-15 09:30',
  updated_at: '2026-06-15 09:30',
  cooking: [
    { ingredients: { name: '玉ねぎ' }, quantity: 1, unit: '個' },
    { ingredients: { name: 'にんじん' }, quantity: 2, unit: '本' },
  ],
  season: [{ seasoning: { name: '醤油' }, quantity: 50, unit: 'ml' }],
  procedure: '材料を切って煮込む。',
  owner: { id: 'u-taro', username: 'taro' },
  shared_user: [],
  title,
  create_time: 30,
  create_for: 2,
  archive_flg: false,
  label: [{ name: '夕食' }],
}))

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
  await page.route('**/api/label/', (route) =>
    route.fulfill({ status: 200, json: [{ name: '夕食' }] }),
  )
  await page.route('**/api/users/', (route) =>
    route.fulfill({ status: 200, json: [{ id: 'u-taro', username: 'taro' }] }),
  )
  await page.route('**/api/recipes/', (route) => route.fulfill({ status: 200, json: recipes }))
}

async function login(page: Page) {
  await page.goto('/')
  await page.fill('#email', 'taro@example.com')
  await page.fill('#password', 'password123')
  await page.getByRole('button', { name: 'ログイン' }).click()
  await page.waitForURL(/\/top$/)
  await page.getByText('カレー').first().waitFor()
}

const viewports = [
  { name: 'mobile-375', width: 375, height: 720 },
  { name: 'tablet-768', width: 768, height: 900 },
]

for (const vp of viewports) {
  test(`screenshot ${vp.name}`, async ({ page }) => {
    await page.setViewportSize({ width: vp.width, height: vp.height })
    await mockApi(page)
    await login(page)

    // 一覧(グリッド折り返し)
    await page.screenshot({ path: `screenshots/${vp.name}-list.png`, fullPage: true })

    // 作成ダイアログ(縦積み・チップ折り返し・dvh 高さ)
    await page.getByRole('button', { name: 'レシピを追加' }).click()
    await page.getByRole('dialog').waitFor({ state: 'visible' })
    // 開閉アニメーション(fade/zoom)が完了してから撮る。
    await page.waitForTimeout(600)
    await page.screenshot({ path: `screenshots/${vp.name}-create-dialog.png` })
  })
}
