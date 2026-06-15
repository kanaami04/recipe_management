import axios from 'axios'
import { http, HttpResponse } from 'msw'
import { afterEach, describe, expect, it } from 'vitest'

import { server } from '@/test/msw/server'

import { attachAuthInterceptors } from './interceptors'
import { requireAuth } from './requireAuth'
import {
  clearAccessToken,
  getAccessToken,
  setAccessToken,
  subscribeAccessToken,
} from './tokenStore'

afterEach(() => clearAccessToken())

describe('tokenStore', () => {
  it('set した access を get で取得できること。', () => {
    // Arrange & Act
    setAccessToken('abc')

    // Assert
    expect(getAccessToken()).toBe('abc')
  })

  it('clear した時、null になること。', () => {
    // Arrange
    setAccessToken('abc')

    // Act
    clearAccessToken()

    // Assert
    expect(getAccessToken()).toBeNull()
  })

  it('購読したリスナーに変更が通知され、解除後は通知されないこと。', () => {
    // Arrange
    const seen: (string | null)[] = []
    const unsubscribe = subscribeAccessToken((t) => seen.push(t))

    // Act
    setAccessToken('x')
    unsubscribe()
    setAccessToken('y')

    // Assert
    expect(seen).toEqual(['x'])
  })
})

describe('attachAuthInterceptors', () => {
  it('401 を受けた時、refresh で更新して元リクエストを再試行し成功すること。', async () => {
    // Arrange
    setAccessToken('old')
    server.use(
      http.post('*/api/token/refresh/', () => HttpResponse.json({ access: 'new' })),
      http.get('*/api/thing', ({ request }) =>
        request.headers.get('Authorization') === 'Bearer new'
          ? HttpResponse.json({ ok: true })
          : new HttpResponse(null, { status: 401 }),
      ),
    )
    const instance = axios.create({ baseURL: '' })
    attachAuthInterceptors(instance)

    // Act
    const res = await instance.get('/api/thing')

    // Assert
    expect(res.status).toBe(200)
  })

  it('同時多発の 401 でも refresh が 1 回だけ走ること(single-flight)。', async () => {
    // Arrange
    setAccessToken('old')
    let refreshCount = 0
    server.use(
      http.post('*/api/token/refresh/', () => {
        refreshCount += 1
        return HttpResponse.json({ access: 'new' })
      }),
      http.get('*/api/thing', ({ request }) =>
        request.headers.get('Authorization') === 'Bearer new'
          ? HttpResponse.json({ ok: true })
          : new HttpResponse(null, { status: 401 }),
      ),
    )
    const instance = axios.create({ baseURL: '' })
    attachAuthInterceptors(instance)

    // Act
    await Promise.all([
      instance.get('/api/thing'),
      instance.get('/api/thing'),
      instance.get('/api/thing'),
    ])

    // Assert
    expect(refreshCount).toBe(1)
  })
})

describe('requireAuth', () => {
  it('access がある時、refresh せず null を返すこと。', async () => {
    // Arrange
    setAccessToken('present')

    // Act
    const result = await requireAuth()

    // Assert
    expect(result).toBeNull()
  })

  it('access が無く refresh が成功した時、access を保持して null を返すこと。', async () => {
    // Arrange
    clearAccessToken()
    server.use(http.post('*/api/token/refresh/', () => HttpResponse.json({ access: 'fresh' })))

    // Act
    const result = await requireAuth()

    // Assert
    expect(result).toBeNull()
    expect(getAccessToken()).toBe('fresh')
  })

  it('access が無く refresh も失敗した時、"/" へリダイレクトを投げること。', async () => {
    // Arrange
    clearAccessToken()
    server.use(http.post('*/api/token/refresh/', () => new HttpResponse(null, { status: 401 })))

    // Act & Assert
    await expect(requireAuth()).rejects.toMatchObject({ status: 302 })
  })
})
