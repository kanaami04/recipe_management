import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'

import { api } from '@/lib/api'

import { server } from './server'

describe('MSW + axios', () => {
  it('登録したハンドラに対し GET した時、モックレスポンスが返ること。', async () => {
    // Arrange
    server.use(http.get('*/api/ping', () => HttpResponse.json({ message: 'pong' })))

    // Act
    const res = await api.get('/api/ping')

    // Assert
    expect(res.data).toEqual({ message: 'pong' })
  })
})
