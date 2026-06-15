import axios from 'axios'
import { http, HttpResponse } from 'msw'
import { describe, expect, it } from 'vitest'

import { server } from './server'

describe('MSW + axios', () => {
  it('登録したハンドラに対し GET した時、モックレスポンスが返ること。', async () => {
    // Arrange
    server.use(http.get('*/api/ping', () => HttpResponse.json({ message: 'pong' })))

    // Act
    const res = await axios.get('/api/ping')

    // Assert
    expect(res.data).toEqual({ message: 'pong' })
  })
})
