import { expect, type Page, test } from '@playwright/test'

const recipe = (id: string, title: string, labels: string[]) => ({
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
  label: labels.map((name) => ({ name })),
})

async function mockApi(page: Page) {
  // ラベルマスタ(id 付き)。作成・改名・削除で書き換える。
  const labels = [{ id: 'l1', name: '和食' }]
  let seq = 1

  await page.route('**/api/token/', (r) => r.fulfill({ status: 200, json: { access: 'a' } }))
  await page.route('**/api/token/refresh/', (r) =>
    r.fulfill({ status: 200, json: { access: 'a' } }),
  )
  await page.route('**/api/user_info/', (r) =>
    r.fulfill({ status: 200, json: { id: 'u-taro', username: 'taro', email: 'taro@example.com' } }),
  )
  await page.route('**/api/users/', (r) => r.fulfill({ status: 200, json: [] }))
  await page.route('**/api/recipes/', (r) =>
    r.fulfill({
      status: 200,
      json: [recipe('r1', 'カレー', ['和食']), recipe('r2', 'サラダ', ['洋食'])],
    }),
  )

  // ラベル個別(改名・削除)
  await page.route('**/api/label/*/', (r) => {
    const id = r
      .request()
      .url()
      .match(/\/api\/label\/([^/]+)\//)?.[1]
    const idx = labels.findIndex((l) => l.id === id)
    if (r.request().method() === 'PUT' && idx >= 0) {
      labels[idx] = { ...labels[idx], name: r.request().postDataJSON().name }
      return r.fulfill({ status: 200, json: labels[idx] })
    }
    if (r.request().method() === 'DELETE' && idx >= 0) {
      labels.splice(idx, 1)
      return r.fulfill({ status: 204, body: '' })
    }
    return r.fallback()
  })

  // ラベル一覧・作成
  await page.route('**/api/label/', (r) => {
    if (r.request().method() === 'POST') {
      seq += 1
      const created = { id: `l${seq}`, name: r.request().postDataJSON().name }
      labels.push(created)
      return r.fulfill({ status: 201, json: created })
    }
    return r.fulfill({ status: 200, json: labels })
  })
}

async function login(page: Page) {
  await page.goto('/')
  await page.fill('#email', 'taro@example.com')
  await page.fill('#password', 'password123')
  await page.getByRole('button', { name: 'ログイン' }).click()
  await expect(page).toHaveURL(/\/top$/)
}

test('ラベルを作成・改名・削除できる', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)

  // Act: サイドバーからラベル管理へ
  await page.getByRole('button', { name: 'ラベル', exact: true }).click()
  await expect(page).toHaveURL(/\/top\/labels$/)
  await expect(page.getByText('和食')).toBeVisible()

  // Act: 新規作成
  await page.getByPlaceholder('新しいラベル名').fill('中華')
  await page.getByRole('button', { name: '追加' }).click()

  // Assert: 追加され一覧に出る
  await expect(page.getByText('ラベルを追加しました')).toBeVisible()
  await expect(page.getByText('中華')).toBeVisible()

  // Act: 「和食」を改名
  await page
    .getByRole('listitem')
    .filter({ hasText: '和食' })
    .getByRole('button', { name: '名前を変更' })
    .click()
  // 編集用の入力欄は li 内(新規作成の入力欄は li の外)。
  await page.locator('li input').fill('日本料理')
  await page.getByRole('button', { name: '保存' }).click()

  // Assert
  await expect(page.getByText('ラベル名を変更しました')).toBeVisible()
  await expect(page.getByText('日本料理')).toBeVisible()

  // Act: 「中華」を削除
  await page
    .getByRole('listitem')
    .filter({ hasText: '中華' })
    .getByRole('button', { name: '削除' })
    .click()
  await page.getByRole('alertdialog').getByRole('button', { name: '削除' }).click()

  // Assert
  await expect(page.getByText('ラベルを削除しました')).toBeVisible()
  await expect(page.getByText('中華')).toHaveCount(0)
})

test('一覧をラベルで絞り込める', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await expect(page.getByText('カレー')).toBeVisible()
  await expect(page.getByText('サラダ')).toBeVisible()

  // Act: 「和食」で絞り込む
  await page.getByRole('button', { name: '和食', exact: true }).click()

  // Assert: 和食のカレーだけ残る
  await expect(page.getByText('カレー')).toBeVisible()
  await expect(page.getByText('サラダ')).toHaveCount(0)

  // Act: 「すべて」で解除
  await page.getByRole('button', { name: 'すべて', exact: true }).click()

  // Assert
  await expect(page.getByText('サラダ')).toBeVisible()
})

test('作成画面のラベル候補が多くてもスクロールできる', async ({ page }) => {
  // Arrange: ラベルを多め(16件)にして候補リストが溢れるようにする
  await mockApi(page)
  const many = Array.from({ length: 16 }, (_, i) => ({ id: `l${i}`, name: `ラベル${i + 1}` }))
  await page.route('**/api/label/', (r) => r.fulfill({ status: 200, json: many }))
  await login(page)

  // Act: 作成ダイアログ → ラベルのポップオーバーを開く
  await page.getByRole('button', { name: 'レシピを追加' }).click()
  await page.getByRole('button', { name: '選択してください' }).first().click()
  const scroller = page.locator('[data-slot="popover-content"] > div').first()
  await expect(scroller).toBeVisible()

  // Act: ポップオーバー内でホイールスクロール(Dialog の scroll-lock に阻まれないこと)
  const box = await scroller.boundingBox()
  if (box) {
    await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2)
    await page.mouse.wheel(0, 400)
  }

  // Assert: 候補リストが実際にスクロールする
  await expect.poll(() => scroller.evaluate((el) => el.scrollTop)).toBeGreaterThan(0)
})
