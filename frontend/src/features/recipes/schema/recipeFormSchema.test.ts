import { describe, expect, it } from 'vitest'

import { recipeFormSchema, toFormValues, toRecipeRequest } from './recipeFormSchema'

const validValues = {
  title: 'カレー',
  createFor: '2',
  createTime: '30',
  procedure: '煮る',
  archiveFlg: false,
  label: ['夕食'],
  sharedUser: ['taro'],
  ingredients: [{ name: '玉ねぎ', quantity: 1, unit: '個' }],
  seasoning: [],
}

describe('recipeFormSchema', () => {
  it('妥当な入力の時、検証を通過すること。', () => {
    // Act
    const result = recipeFormSchema.safeParse(validValues)

    // Assert
    expect(result.success).toBe(true)
  })

  it('タイトルが空の時、検証に失敗すること。', () => {
    // Act
    const result = recipeFormSchema.safeParse({ ...validValues, title: '' })

    // Assert
    expect(result.success).toBe(false)
  })

  it('食材が空の時、検証に失敗すること。', () => {
    // Act
    const result = recipeFormSchema.safeParse({ ...validValues, ingredients: [] })

    // Assert
    expect(result.success).toBe(false)
  })
})

describe('toRecipeRequest', () => {
  it('create_time が空文字の時、null に変換されること。', () => {
    // Act
    const req = toRecipeRequest({ ...validValues, createTime: '' })

    // Assert
    expect(req.create_time).toBeNull()
  })

  it('createFor の文字列が数値に変換され、API 形へネストされること。', () => {
    // Act
    const req = toRecipeRequest(validValues)

    // Assert
    expect(req.create_for).toBe(2)
    expect(req.cooking).toEqual([{ ingredients: { name: '玉ねぎ' }, quantity: 1, unit: '個' }])
  })
})

describe('toFormValues', () => {
  it('create_time が null のレスポンスの時、空文字に変換されること。', () => {
    // Arrange
    const recipe = {
      id: 'r1',
      created_at: '2026-06-15 09:30',
      updated_at: '2026-06-15 09:30',
      cooking: [],
      season: [],
      procedure: '',
      owner: { id: 'u-taro', username: 'taro' },
      shared_user: [],
      title: 'カレー',
      create_time: null,
      create_for: 2,
      archive_flg: false,
      label: [],
    }

    // Act
    const values = toFormValues(recipe)

    // Assert
    expect(values.createTime).toBe('')
    expect(values.createFor).toBe('2')
  })
})
