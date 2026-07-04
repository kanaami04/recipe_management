import { expect, type Page, test } from '@playwright/test'

async function mockApi(page: Page) {
  const user = {
    id: 'u-taro',
    username: 'taro',
    email: 'taro@example.com',
    created_at: '2026-06-15 09:30',
  }
  await page.route('**/api/token/', (r) => r.fulfill({ status: 200, json: { access: 'a' } }))
  await page.route('**/api/token/refresh/', (r) =>
    r.fulfill({ status: 200, json: { access: 'a' } }),
  )
  await page.route('**/api/auth/logout/', (r) => r.fulfill({ status: 204, body: '' }))
  await page.route('**/api/users/', (r) => r.fulfill({ status: 200, json: [] }))
  await page.route('**/api/label/', (r) => r.fulfill({ status: 200, json: [] }))
  await page.route('**/api/recipes/', (r) => r.fulfill({ status: 200, json: [] }))

  await page.route('**/api/user_info/password/', (r) =>
    r.request().method() === 'PUT' ? r.fulfill({ status: 204, body: '' }) : r.fallback(),
  )
  // メール変更(パスワード確認)。パスワードが一致すれば email を反映する。
  await page.route('**/api/user_info/email/', (r) => {
    if (r.request().method() !== 'PUT') return r.fallback()
    const body = r.request().postDataJSON()
    if (body.password !== 'password123') {
      return r.fulfill({ status: 400, json: { message: 'wrong password' } })
    }
    user.email = body.email
    return r.fulfill({ status: 200, json: user })
  })
  await page.route('**/api/user_info/', (r) => {
    const method = r.request().method()
    if (method === 'PUT') {
      user.username = r.request().postDataJSON().username
      return r.fulfill({ status: 200, json: user })
    }
    if (method === 'DELETE') {
      return r.fulfill({ status: 204, body: '' })
    }
    return r.fulfill({ status: 200, json: user })
  })
}

async function login(page: Page) {
  await page.goto('/')
  await page.fill('#email', 'taro@example.com')
  await page.fill('#password', 'password123')
  await page.getByRole('button', { name: 'ログイン' }).click()
  await expect(page).toHaveURL(/\/top$/)
}

test('ユーザーメニューからアカウント画面へ遷移できる', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)

  // Act: サイドバーのユーザー → アカウント
  await page.getByText('taro', { exact: true }).click()
  await page.getByRole('menuitem', { name: 'アカウント' }).click()

  // Assert
  await expect(page).toHaveURL(/\/top\/account$/)
  await expect(page.getByText('登録日: 2026-06-15 09:30')).toBeVisible()
})

test('スマホでアカウントへ遷移するとサイドバーが閉じる', async ({ page }) => {
  // Arrange: モバイル幅(サイドバーが Sheet)でログイン
  await page.setViewportSize({ width: 375, height: 800 })
  await mockApi(page)
  await login(page)

  // Act: サイドバーを開き、ユーザー → アカウント
  await page.getByRole('button', { name: 'Toggle Sidebar' }).click()
  await page.getByText('taro', { exact: true }).click()
  await page.getByRole('menuitem', { name: 'アカウント' }).click()

  // Assert: 遷移し、サイドバー(Sheet=dialog)が閉じる
  await expect(page).toHaveURL(/\/top\/account$/)
  await expect(page.getByRole('dialog')).toHaveCount(0)
})

test('プロフィール(ユーザー名)を更新できる', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.goto('/top/account')
  await expect(page.locator('#username')).toHaveValue('taro')

  // Act
  await page.locator('#username').fill('taro2')
  await page.getByRole('button', { name: '保存' }).click()

  // Assert
  await expect(page.getByText('プロフィールを更新しました')).toBeVisible()
})

test('プロフィールのメール欄は編集できない(disabled)', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.goto('/top/account')

  // Assert: プロフィールのメール欄は現在値を表示しつつ無効化されている
  await expect(page.locator('#email')).toHaveValue('taro@example.com')
  await expect(page.locator('#email')).toBeDisabled()
})

test('パスワード確認のうえメールアドレスを変更できる', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.goto('/top/account')

  // Act: 新しいメール + 現在のパスワードで変更
  await page.locator('#newEmail').fill('taro2@example.com')
  await page.locator('#emailPassword').fill('password123')
  await page.getByRole('button', { name: 'メールアドレスを変更' }).click()

  // Assert
  await expect(page.getByText('メールアドレスを変更しました')).toBeVisible()
})

test('パスワードが違うとメールアドレスを変更できない', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.goto('/top/account')

  // Act: 誤ったパスワードで変更
  await page.locator('#newEmail').fill('taro2@example.com')
  await page.locator('#emailPassword').fill('wrongpass')
  await page.getByRole('button', { name: 'メールアドレスを変更' }).click()

  // Assert
  await expect(page.getByText('パスワードが違います')).toBeVisible()
})

test('パスワードを変更できる', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.goto('/top/account')

  // Act
  await page.locator('#currentPassword').fill('password123')
  await page.locator('#newPassword').fill('newpassword123')
  await page.locator('#confirmPassword').fill('newpassword123')
  await page.getByRole('button', { name: 'パスワードを変更' }).click()

  // Assert
  await expect(page.getByText('パスワードを変更しました')).toBeVisible()
})

test('確認用パスワードが一致しないと検証エラーになる', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.goto('/top/account')

  // Act: 確認用を別の値にして送信
  await page.locator('#currentPassword').fill('password123')
  await page.locator('#newPassword').fill('newpassword123')
  await page.locator('#confirmPassword').fill('different123')
  await page.getByRole('button', { name: 'パスワードを変更' }).click()

  // Assert
  await expect(page.getByText('パスワードが一致しません')).toBeVisible()
})

test('アカウントを削除するとログイン画面に戻る', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.goto('/top/account')

  // Act: 削除 → 確認ダイアログで確定
  await page.getByRole('button', { name: 'アカウントを削除' }).click()
  await page.getByRole('alertdialog').getByRole('button', { name: '削除' }).click()

  // Assert: ログイン画面へ戻り、トーストが出る
  await expect(page).toHaveURL(/\/$/)
  await expect(page.getByText('アカウントを削除しました')).toBeVisible()
})
