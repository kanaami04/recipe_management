import { expect, test } from '@playwright/test'

// PWA の検証。SW の登録を確認するため、このファイルだけ
// playwright.config.ts の serviceWorkers: 'block' を上書きする。
test.use({ serviceWorkers: 'allow' })

test('Service Worker が登録される', async ({ page }) => {
  // Arrange & Act
  await page.goto('/')

  // Assert: SW が activate されるまで待つ
  const scope = await page.evaluate(async () => {
    const registration = await navigator.serviceWorker.ready
    return registration.scope
  })
  expect(scope).toBe('http://localhost:4173/')
})

test('Web App Manifest が配信され、head から参照されている', async ({ page, request }) => {
  // Arrange & Act
  await page.goto('/')

  // Assert: <link rel="manifest"> が存在する
  await expect(page.locator('link[rel="manifest"]')).toHaveAttribute(
    'href',
    '/manifest.webmanifest',
  )

  // Assert: manifest 本体が配信され、インストールに必要な項目を含む
  const res = await request.get('/manifest.webmanifest')
  expect(res.status()).toBe(200)
  const manifest = await res.json()
  expect(manifest.name).toBe('cookience')
  expect(manifest.display).toBe('standalone')
  expect(manifest.icons.length).toBeGreaterThanOrEqual(3)
})
