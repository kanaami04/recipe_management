import { expect, type Page, test } from '@playwright/test'

const owner = { id: 'u-taro', username: 'taro', avatar_url: null }

// 買い物リストを可変の状態で持ち、追加/チェック/削除/一括削除を反映するモック。
// サーバは常に更新後のリスト全体(ShoppingListResponse)を返す。
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
  await page.route('**/api/users/', (route) =>
    route.fulfill({
      status: 200,
      json: [{ id: 'u-hanako', username: 'hanako', avatar_url: null }],
    }),
  )
  await page.route('**/api/recipes/', (route) => route.fulfill({ status: 200, json: [] }))

  const listId = 'sl1'
  let seq = 0
  const items: { id: string; name: string; checked: boolean }[] = []
  const body = () => ({ id: listId, owner, shared_user: [], items })

  // 一括削除(checked)を item/:id より先に評価させるため、より具体的なパスを後に登録する
  // (Playwright は最後に登録したルートを優先する)。
  await page.route('**/api/shopping_list/', (route) => route.fulfill({ status: 200, json: body() }))
  await page.route('**/api/shopping_list/*/items/', (route) => {
    const payload = route.request().postDataJSON() as { name: string }
    items.push({ id: `i${++seq}`, name: payload.name, checked: false })
    return route.fulfill({ status: 200, json: body() })
  })
  await page.route('**/api/shopping_list/*/items/*/', (route) => {
    const method = route.request().method()
    const itemId = route
      .request()
      .url()
      .match(/\/items\/([^/]+)\//)?.[1]
    const index = items.findIndex((it) => it.id === itemId)
    if (method === 'PUT') {
      const payload = route.request().postDataJSON() as { checked: boolean }
      if (index >= 0) items[index].checked = payload.checked
    } else if (method === 'DELETE' && index >= 0) {
      items.splice(index, 1)
    }
    return route.fulfill({ status: 200, json: body() })
  })
  await page.route('**/api/shopping_list/*/items/checked/', (route) => {
    for (let i = items.length - 1; i >= 0; i--) if (items[i].checked) items.splice(i, 1)
    return route.fulfill({ status: 200, json: body() })
  })
  // reorder は :item_id より後に登録し、Playwright の「後勝ち」で優先させる。
  await page.route('**/api/shopping_list/*/items/reorder/', (route) => {
    const payload = route.request().postDataJSON() as { item_ids: string[] }
    const byId = new Map(items.map((it) => [it.id, it]))
    const reordered = payload.item_ids
      .map((id) => byId.get(id))
      .filter((it): it is (typeof items)[number] => it != null)
    items.splice(0, items.length, ...reordered)
    return route.fulfill({ status: 200, json: body() })
  })
}

async function login(page: Page) {
  await page.goto('/')
  await page.fill('#email', 'taro@example.com')
  await page.fill('#password', 'password123')
  await page.getByRole('button', { name: 'ログイン' }).click()
  await expect(page).toHaveURL(/\/top$/)
}

test('サイドバーから買い物リストへ遷移し、追加・チェック・一括削除ができる', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)

  // Act: サイドバーの「買い物リスト」タブへ遷移
  await page.getByRole('button', { name: '買い物リスト' }).click()
  await expect(page).toHaveURL(/\/top\/shopping-list$/)

  // 品物を 3 つ追加(Enter 連続追加)
  const input = page.getByPlaceholder('追加する品物')
  for (const name of ['牛乳', '卵', 'パン']) {
    await input.fill(name)
    await input.press('Enter')
    await expect(page.getByText(name)).toBeVisible()
  }

  // 先頭の「牛乳」をチェック → 取り消し線 + 下部へ沈む
  const milkRow = page.locator('li', { hasText: '牛乳' })
  await milkRow.getByRole('checkbox').click()
  await expect(page.getByText('牛乳')).toHaveClass(/line-through/)

  // Assert: チェック済みが一番下に来る
  const names = await page.locator('li span.truncate').allInnerTexts()
  expect(names[names.length - 1]).toBe('牛乳')

  // 一括削除
  await page.getByRole('button', { name: /チェック済みを削除/ }).click()
  await page.getByRole('alertdialog').getByRole('button', { name: '削除' }).click()
  await expect(page.getByText('チェック済みを削除しました')).toBeVisible()
  await expect(page.getByText('牛乳')).toHaveCount(0)
})

test('グリップをドラッグして未チェック項目を並び替えできる', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.getByRole('button', { name: '買い物リスト' }).click()
  const input = page.getByPlaceholder('追加する品物')
  for (const name of ['牛乳', '卵', 'パン']) {
    await input.fill(name)
    await input.press('Enter')
    await expect(page.getByText(name)).toBeVisible()
  }

  // Act: 先頭「牛乳」のグリップを掴み、3 番目「パン」の位置までドラッグする
  const milkGrip = page.locator('li', { hasText: '牛乳' }).getByRole('button', { name: '並び替え' })
  const breadRow = page.locator('li', { hasText: 'パン' })
  const from = await milkGrip.boundingBox()
  const to = await breadRow.boundingBox()
  if (!from || !to) throw new Error('bounding box not found')
  await page.mouse.move(from.x + from.width / 2, from.y + from.height / 2)
  await page.mouse.down()
  // MouseSensor は 8px 動いて初めて起動するので、まず小さく動かしてから目標へ運ぶ。
  await page.mouse.move(from.x + from.width / 2, from.y + from.height / 2 + 12)
  await page.mouse.move(to.x + to.width / 2, to.y + to.height / 2, { steps: 6 })
  await page.mouse.up()

  // Assert: 「牛乳」が末尾へ移動する
  await expect(async () => {
    const names = await page.locator('li span.truncate').allInnerTexts()
    expect(names[names.length - 1]).toBe('牛乳')
  }).toPass()
})

test('見た目確認: 買い物リスト(デスクトップ / モバイル)', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.getByRole('button', { name: '買い物リスト' }).click()
  const input = page.getByPlaceholder('追加する品物')
  for (const name of ['牛乳', '卵', 'パン', 'トマト']) {
    await input.fill(name)
    await input.press('Enter')
    await expect(page.getByText(name)).toBeVisible()
  }
  await page.locator('li', { hasText: '卵' }).getByRole('checkbox').click()
  await expect(page.getByText('卵')).toHaveClass(/line-through/)

  // Act: デスクトップ幅で撮影(screenshots/ は .gitignore 済み)
  await page.screenshot({ path: 'screenshots/shopping-list-desktop.png', fullPage: true })

  // モバイル幅で撮影
  await page.setViewportSize({ width: 375, height: 800 })
  await page.screenshot({ path: 'screenshots/shopping-list-mobile.png', fullPage: true })
})
