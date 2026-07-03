import { expect, type Page, test } from '@playwright/test'

// 一覧表示・作成レスポンスに使うレシピ(API 契約 RecipeResponse の形)。
const recipe = {
  id: 'r1',
  created_at: '2026-06-15 09:30',
  updated_at: '2026-06-15 09:30',
  cooking: [{ ingredients: { name: '玉ねぎ' }, quantity: 1, unit: '個' }],
  season: [],
  procedure: '煮る',
  owner: { id: 'u-taro', username: 'taro' },
  shared_user: [],
  title: 'カレー',
  create_time: 30,
  create_for: 2,
  archive_flg: false,
  label: [{ name: '夕食' }],
}

// API をブラウザ側のルートモックで差し替える(バックエンド/DB 不要)。
async function mockApi(page: Page) {
  await page.route('**/api/token/', (route) =>
    route.fulfill({ status: 200, json: { access: 'fake-access' } }),
  )
  await page.route('**/api/token/refresh/', (route) =>
    route.fulfill({ status: 200, json: { access: 'fake-access' } }),
  )
  await page.route('**/api/auth/register/', (route) =>
    route.fulfill({
      status: 201,
      json: { id: 'u-hanako', username: 'hanako', email: 'hanako@example.com' },
    }),
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
  await page.route('**/api/auth/logout/', (route) => route.fulfill({ status: 204, body: '' }))
  // レシピはミュータブルなリストで持ち、作成/更新/削除を一覧に反映する。
  const recipes = [{ ...recipe }]

  // 個別レシピ(id 付きパス)の更新・削除。一覧のリストを書き換える。
  await page.route('**/api/recipes/*/', (route) => {
    const method = route.request().method()
    const id = route
      .request()
      .url()
      .match(/\/api\/recipes\/([^/]+)\//)?.[1]
    const index = recipes.findIndex((r) => r.id === id)
    if (method === 'PUT') {
      const body = route.request().postDataJSON()
      // スカラー項目だけレスポンス形へ反映する(タイトル等の反映確認に使う)。
      recipes[index] = {
        ...recipes[index],
        title: body.title,
        create_for: body.create_for,
        create_time: body.create_time,
        procedure: body.procedure,
        archive_flg: body.archive_flg,
      }
      return route.fulfill({ status: 200, json: recipes[index] })
    }
    if (method === 'DELETE') {
      recipes.splice(index, 1)
      return route.fulfill({ status: 204, body: '' })
    }
    return route.fallback()
  })

  await page.route('**/api/recipes/', (route) => {
    if (route.request().method() === 'POST') {
      const created = { ...recipe, id: 'r2', title: '新レシピ' }
      recipes.push(created)
      return route.fulfill({ status: 201, json: created })
    }
    return route.fulfill({ status: 200, json: recipes })
  })
}

async function login(page: Page) {
  await page.goto('/')
  await page.fill('#email', 'taro@example.com')
  await page.fill('#password', 'password123')
  await page.getByRole('button', { name: 'ログイン' }).click()
  await expect(page).toHaveURL(/\/top$/)
}

test('ログインするとレシピ一覧が表示される', async ({ page }) => {
  // Arrange
  await mockApi(page)

  // Act
  await login(page)

  // Assert
  await expect(page.getByText('カレー')).toBeVisible()
})

test('サインアップするとそのままログインして一覧へ遷移する', async ({ page }) => {
  // Arrange
  await mockApi(page)

  // Act: サインアップ画面で登録 → 自動ログイン
  await page.goto('/signup')
  await page.fill('#username', 'hanako')
  await page.fill('#email', 'hanako@example.com')
  await page.fill('#password', 'password123')
  await page.getByRole('button', { name: '登録' }).click()

  // Assert
  await expect(page).toHaveURL(/\/top$/)
  await expect(page.getByText('カレー')).toBeVisible()
})

test('レシピを新規作成すると成功トーストが出る', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)

  // Act: 作成ダイアログを開いて最小限の項目を入力し送信する
  await page.getByRole('button', { name: 'レシピを追加' }).click()
  await page.getByPlaceholder('タイトル').fill('新レシピ')
  await page.getByRole('combobox').first().click() // 人数
  await page.getByRole('option', { name: '2', exact: true }).click()
  // 食材・調味料それぞれ初期1行の name と単位を埋める(名前・単位は必須のため)。
  // 単位を選ぶと数量はその単位の既定値に入る。個は食材行、大さじは調味料行にのみ現れ一意。
  await page.getByPlaceholder('名前').nth(0).fill('じゃがいも')
  await page.getByRole('button', { name: '個', exact: true }).click()
  await page.getByPlaceholder('名前').nth(1).fill('塩')
  await page.getByRole('button', { name: '大さじ', exact: true }).click()
  await page.getByRole('button', { name: '作成' }).click()

  // Assert
  await expect(page.getByText('レシピを作成しました')).toBeVisible()
})

test('調味料を入力しなくてもレシピを作成できる', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)

  // Act: 食材だけ埋めて作成する(調味料の初期空行は触らない)
  await page.getByRole('button', { name: 'レシピを追加' }).click()
  await page.getByPlaceholder('タイトル').fill('調味料なしレシピ')
  await page.getByRole('combobox').first().click() // 人数
  await page.getByRole('option', { name: '2', exact: true }).click()
  await page.getByPlaceholder('名前').first().fill('じゃがいも')
  await page.getByRole('button', { name: '個', exact: true }).click()
  await page.getByRole('button', { name: '作成' }).click()

  // Assert: 空の調味料行に阻まれず作成できる
  await expect(page.getByText('レシピを作成しました')).toBeVisible()
})

test('食材を入力せず作成すると警告が出て作成されない', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)

  // Act: タイトル・人数だけ入れて作成を押す
  await page.getByRole('button', { name: 'レシピを追加' }).click()
  await page.getByPlaceholder('タイトル').fill('食材なしレシピ')
  await page.getByRole('combobox').first().click() // 人数
  await page.getByRole('option', { name: '2', exact: true }).click()
  await page.getByRole('button', { name: '作成' }).click()

  // Assert: 必須警告が表示される
  await expect(page.getByText('食材は1つ以上必要です')).toBeVisible()
})

test('レシピを編集して更新すると詳細に反映される', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)

  // Act: 一覧のカードを開き、⋮ メニューから編集してタイトルを変えて更新する
  await page.getByText('カレー').click()
  await page.getByRole('button', { name: '操作メニュー' }).click()
  await page.getByRole('menuitem', { name: '編集' }).click()
  const title = page.getByPlaceholder('タイトル')
  await title.fill('ビーフカレー')
  await page.getByRole('button', { name: '更新' }).click()

  // Assert: 成功トーストが出て、詳細に新しいタイトルが反映される
  await expect(page.getByText('レシピを編集しました')).toBeVisible()
  await expect(page.getByText('ビーフカレー', { exact: true }).first()).toBeVisible()
})

test('レシピを削除すると一覧から消える', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)

  // Act: 詳細を開き、⋮ メニューから削除 → 確認ダイアログで削除を確定する
  await page.getByText('カレー').click()
  await page.getByRole('button', { name: '操作メニュー' }).click()
  await page.getByRole('menuitem', { name: '削除' }).click()
  await page.getByRole('alertdialog').getByRole('button', { name: '削除' }).click()

  // Assert: 成功トーストが出て、一覧からカレーが消える
  await expect(page.getByText('レシピを削除しました')).toBeVisible()
  await expect(page.getByText('カレー')).toHaveCount(0)
})

test('スマホでログアウトした後もログイン画面を操作できる', async ({ page }) => {
  // Arrange: モバイル幅(サイドバーが Sheet になる)でログインする
  await page.setViewportSize({ width: 375, height: 800 })
  await mockApi(page)
  await login(page)

  // Act: サイドバー → ユーザーメニュー → ログアウト
  await page.getByRole('button', { name: 'Toggle Sidebar' }).click()
  await page.getByText('taro', { exact: true }).click()
  await page.getByRole('menuitem', { name: 'ログアウト' }).click()
  await page.waitForURL('**/')

  // Assert: body の pointer-events ロックが残らず、実際にクリックで操作できる
  // (fill はヒットテストを行わないため、ロック検出には click が必要)
  await page.locator('#email').click()
  await page.locator('#email').fill('again@example.com')
  await expect(page.locator('#email')).toHaveValue('again@example.com')
})

test('認証済みで未定義パスへ行くとログインではなく /top へ戻る', async ({ page }) => {
  // Arrange
  await mockApi(page)
  await login(page)

  // Act: 存在しないパス(sidebar の label/archive 相当)へ遷移する
  await page.goto('/top/archive')

  // Assert: ログインではなく一覧へリダイレクトされる
  await expect(page).toHaveURL(/\/top$/)
})
