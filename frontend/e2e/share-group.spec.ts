import { expect, type Page, test } from '@playwright/test'

const taro = { id: 'u-taro', username: 'taro', avatar_url: null }
const hanako = { id: 'u-hanako', username: 'hanako', avatar_url: null }

type Group = {
  id: string
  name: string
  owner: typeof taro
  members: (typeof taro)[]
  invite_code: string
  invite_code_expires_at: string
  is_owner: boolean
}

// シェアグループの状態(作成/参加で 404 → グループへ遷移する)をルートモックで再現する。
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
  await page.route('**/api/recipes/', (route) => route.fulfill({ status: 200, json: [] }))

  let group: Group | null = null

  await page.route('**/api/share_group/', (route) => {
    const method = route.request().method()
    if (method === 'GET') {
      return group
        ? route.fulfill({ status: 200, json: group })
        : route.fulfill({ status: 404, json: { message: 'not found' } })
    }
    if (method === 'POST') {
      const body = route.request().postDataJSON() as { name?: string }
      group = {
        id: 'g1',
        name: body.name?.trim() || 'マイグループ',
        owner: taro,
        members: [taro],
        invite_code: 'ABCD2345',
        invite_code_expires_at: '2026-07-12 10:00',
        is_owner: true,
      }
      return route.fulfill({ status: 201, json: group })
    }
    return route.fallback()
  })
  // join / leave はより具体的なパス。後に登録して優先させる。
  await page.route('**/api/share_group/join/', (route) => {
    group = {
      id: 'g1',
      name: '我が家',
      owner: hanako,
      members: [hanako, taro],
      invite_code: 'ABCD2345',
      invite_code_expires_at: '2026-07-12 10:00',
      is_owner: false,
    }
    return route.fulfill({ status: 200, json: group })
  })
  await page.route('**/api/share_group/leave/', (route) => {
    group = null
    return route.fulfill({ status: 204, body: '' })
  })
}

async function login(page: Page) {
  await page.goto('/')
  await page.fill('#email', 'taro@example.com')
  await page.fill('#password', 'password123')
  await page.getByRole('button', { name: 'ログイン' }).click()
  await expect(page).toHaveURL(/\/top$/)
}

test('未所属ならオンボーディングが出て、グループを作成できる', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)

  // Act: サイドバーから共有グループへ
  await page.getByRole('button', { name: '共有グループ' }).click()
  await expect(page).toHaveURL(/\/top\/share-group$/)
  await expect(page.getByRole('heading', { name: 'グループを作成' })).toBeVisible()

  // グループ名を入れて作成
  await page.getByPlaceholder('グループ名(例: 我が家)').fill('我が家')
  await page.getByRole('button', { name: '作成' }).click()

  // Assert: 作成後、メンバーと招待コードが表示される
  await expect(page.getByText('共有グループを作成しました')).toBeVisible()
  await expect(page.getByText('ABCD2345')).toBeVisible()
  await expect(page.getByText('メンバー(1)')).toBeVisible()
})

test('招待コードで既存グループに参加できる', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.getByRole('button', { name: '共有グループ' }).click()

  // Act
  await page.getByPlaceholder('招待コード').fill('ABCD2345')
  await page.getByRole('button', { name: '参加' }).click()

  // Assert: 参加後、2 名のメンバーが見える
  await expect(page.getByText('共有グループに参加しました')).toBeVisible()
  await expect(page.getByText('メンバー(2)')).toBeVisible()
})

test('グループを解散するとリロードせずオンボーディングへ戻る', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.getByRole('button', { name: '共有グループ' }).click()
  await page.getByPlaceholder('グループ名(例: 我が家)').fill('我が家')
  await page.getByRole('button', { name: '作成' }).click()
  await expect(page.getByText('ABCD2345')).toBeVisible()

  // Act: 解散する
  await page.getByRole('button', { name: 'グループを解散' }).click()
  await page.getByRole('alertdialog').getByRole('button', { name: '解散' }).click()

  // Assert: リロードなしで作成/参加のオンボーディングへ戻る
  await expect(page.getByText('グループを解散しました')).toBeVisible()
  await expect(page.getByRole('heading', { name: 'グループを作成' })).toBeVisible()
})

test('見た目確認: 共有グループ(オンボーディング / グループ)', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.getByRole('button', { name: '共有グループ' }).click()
  await expect(page.getByRole('heading', { name: 'グループを作成' })).toBeVisible()

  // オンボーディング(デスクトップ)
  await page.screenshot({ path: 'screenshots/share-group-onboarding.png', fullPage: true })

  // 作成後のグループ表示(デスクトップ)
  await page.getByPlaceholder('グループ名(例: 我が家)').fill('我が家')
  await page.getByRole('button', { name: '作成' }).click()
  await expect(page.getByText('ABCD2345')).toBeVisible()
  await page.screenshot({ path: 'screenshots/share-group-detail.png', fullPage: true })

  // モバイル幅
  await page.setViewportSize({ width: 375, height: 800 })
  await page.screenshot({ path: 'screenshots/share-group-mobile.png', fullPage: true })
})
