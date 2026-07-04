import { expect, type Page, test } from '@playwright/test'

async function mockApi(page: Page) {
  const user: {
    id: string
    username: string
    email: string
    created_at: string
    avatar_url: string | null
  } = {
    id: 'u-taro',
    username: 'taro',
    email: 'taro@example.com',
    created_at: '2026-06-15 09:30',
    avatar_url: null,
  }
  await page.route('**/api/token/', (r) => r.fulfill({ status: 200, json: { access: 'a' } }))
  await page.route('**/api/token/refresh/', (r) =>
    r.fulfill({ status: 200, json: { access: 'a' } }),
  )
  await page.route('**/api/auth/logout/', (r) => r.fulfill({ status: 204, body: '' }))
  await page.route('**/api/users/', (r) => r.fulfill({ status: 200, json: [] }))
  await page.route('**/api/label/', (r) => r.fulfill({ status: 200, json: [] }))
  await page.route('**/api/recipes/', (r) => r.fulfill({ status: 200, json: [] }))

  // 署名付き URL への直 PUT(実体は S3)。モックでは 200 を返すだけ。
  await page.route('**/mock-upload', (r) =>
    r.request().method() === 'PUT' ? r.fulfill({ status: 200, body: '' }) : r.fallback(),
  )
  // アバター: 発行(POST)→ 確定(PUT で avatar_url 反映)→ 削除(DELETE で null)。
  await page.route('**/api/user_info/avatar/', (r) => {
    const method = r.request().method()
    if (method === 'POST') {
      return r.fulfill({
        status: 200,
        json: { upload_url: 'http://localhost:9999/mock-upload', key: 'avatars/u-taro/new' },
      })
    }
    if (method === 'PUT') {
      user.avatar_url = 'http://localhost:9999/avatar.png'
      return r.fulfill({ status: 200, json: user })
    }
    if (method === 'DELETE') {
      user.avatar_url = null
      return r.fulfill({ status: 200, json: user })
    }
    return r.fallback()
  })

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

test('「メールアドレスを変更」ボタンで専用画面へ遷移する', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.goto('/top/account')

  // Act
  await page.getByRole('button', { name: 'メールアドレスを変更' }).click()

  // Assert
  await expect(page).toHaveURL(/\/top\/account\/email$/)
  await expect(page.locator('#newEmail')).toBeVisible()
})

test('メール変更画面の「アカウントに戻る」で戻れる', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.goto('/top/account/email')

  // Act
  await page.getByRole('button', { name: 'アカウントに戻る' }).click()

  // Assert
  await expect(page).toHaveURL(/\/top\/account$/)
})

test('パスワード確認のうえメールアドレスを変更するとログイン画面に戻る', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.goto('/top/account/email')

  // Act: 新しいメール + 現在のパスワードで変更
  await page.locator('#newEmail').fill('taro2@example.com')
  await page.locator('#emailPassword').fill('password123')
  await page.getByRole('button', { name: 'メールアドレスを変更' }).click()

  // Assert: 再ログインを促され、ログイン画面へ戻る
  await expect(page).toHaveURL(/\/$/)
  await expect(
    page.getByText('メールアドレスを変更しました。新しいメールアドレスでログインしてください。'),
  ).toBeVisible()
})

test('パスワードが違うとメールアドレスを変更できない', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.goto('/top/account/email')

  // Act: 誤ったパスワードで変更
  await page.locator('#newEmail').fill('taro2@example.com')
  await page.locator('#emailPassword').fill('wrongpass')
  await page.getByRole('button', { name: 'メールアドレスを変更' }).click()

  // Assert: 画面遷移せずエラーが出る
  await expect(page.getByText('パスワードが違います')).toBeVisible()
  await expect(page).toHaveURL(/\/top\/account\/email$/)
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

test('プロフィール画像をアップロードして削除できる', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.goto('/top/account')

  // Act: 画像を選択(隠し input にファイルをセット)→ アップロード
  await page.setInputFiles('input[type="file"]', {
    name: 'avatar.png',
    mimeType: 'image/png',
    buffer: Buffer.from([0x89, 0x50, 0x4e, 0x47]),
  })

  // Assert: 変更成功のトーストが出て、削除ボタンが現れる
  await expect(page.getByText('プロフィール画像を変更しました')).toBeVisible()
  const del = page.getByRole('button', { name: '削除', exact: true })
  await expect(del).toBeVisible()

  // Act: 削除
  await del.click()

  // Assert
  await expect(page.getByText('プロフィール画像を削除しました')).toBeVisible()
})

test('対応外の画像形式はアップロードできない', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)
  await page.goto('/top/account')

  // Act: GIF を選択
  await page.setInputFiles('input[type="file"]', {
    name: 'avatar.gif',
    mimeType: 'image/gif',
    buffer: Buffer.from([0x47, 0x49, 0x46]),
  })

  // Assert: 形式エラーのトーストが出る
  await expect(page.getByText('対応している形式は JPEG / PNG / WebP です')).toBeVisible()
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
