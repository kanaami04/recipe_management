import { describe, expect, it } from 'vitest'

import { zRecipeResponse, zTokenResponse } from './generated/zod.gen'

// 生成パイプライン(型・zod)の疎通確認 (frontend ADR-0007)。
// 本格的な API モックテストは各 feature の実装段階で追加する。
describe('生成された zod スキーマ', () => {
  it('正しい形のレスポンスを渡した時、検証を通過すること。', () => {
    // Arrange(refresh は Cookie 化され body に含まれない, api ADR-0004)
    const valid = { access: 'access-token' }

    // Act
    const result = zTokenResponse.safeParse(valid)

    // Assert
    expect(result.success).toBe(true)
  })

  it('必須項目(access)が欠けたレスポンスを渡した時、検証に失敗すること。', () => {
    // Arrange
    const invalid = {}

    // Act
    const result = zTokenResponse.safeParse(invalid)

    // Assert
    expect(result.success).toBe(false)
  })

  it('create_time が null のレシピを渡した時、検証を通過すること。', () => {
    // Arrange
    const recipe = {
      id: 1,
      created_at: '2026-06-15 09:30',
      updated_at: '2026-06-15 09:30',
      cooking: [],
      season: [],
      procedure: '',
      owner: { id: 1, username: 'taro' },
      shared_user: [],
      title: 'カレー',
      create_time: null,
      create_for: 2,
      archive_flg: false,
      label: [],
    }

    // Act
    const result = zRecipeResponse.safeParse(recipe)

    // Assert
    expect(result.success).toBe(true)
  })
})
