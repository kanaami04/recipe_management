import { describe, expect, it } from 'vitest'

import { recipeFormSchema, toFormValues, toRecipeRequest } from './recipeFormSchema'

const validValues = {
  inputMode: 'manual' as const,
  title: 'カレー',
  createFor: '2',
  createTime: '30',
  procedure: '煮る',
  sourceUrl: '',
  thumbnailUrl: '',
  label: ['夕食'],
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

  it('調味料が空行だけの時、未使用として無視され検証を通過すること。', () => {
    // Act
    const result = recipeFormSchema.safeParse({
      ...validValues,
      seasoning: [{ name: '', quantity: 0, unit: '' }],
    })

    // Assert
    expect(result.success).toBe(true)
  })

  it('食材が空行だけの時、完全な行が無く検証に失敗すること。', () => {
    // Act
    const result = recipeFormSchema.safeParse({
      ...validValues,
      ingredients: [{ name: '', quantity: 0, unit: '' }],
    })

    // Assert
    expect(result.success).toBe(false)
  })

  it('食材の名前だけ入力し単位未選択の時、検証に失敗すること。', () => {
    // Act
    const result = recipeFormSchema.safeParse({
      ...validValues,
      ingredients: [{ name: '玉ねぎ', quantity: 0, unit: '' }],
    })

    // Assert
    expect(result.success).toBe(false)
  })

  it('url モードで参考 URL があれば、タイトル・食材が空でも検証を通過すること。', () => {
    // Act
    const result = recipeFormSchema.safeParse({
      ...validValues,
      inputMode: 'url',
      title: '',
      createFor: '',
      sourceUrl: 'https://www.kurashiru.com/recipes/abc',
      ingredients: [{ name: '', quantity: 0, unit: '' }],
    })

    // Assert
    expect(result.success).toBe(true)
  })

  it('url モードで参考 URL が空の時、検証に失敗すること。', () => {
    // Act
    const result = recipeFormSchema.safeParse({
      ...validValues,
      inputMode: 'url',
      title: '',
      sourceUrl: '',
      ingredients: [{ name: '', quantity: 0, unit: '' }],
    })

    // Assert
    expect(result.success).toBe(false)
  })

  it('url モードで scheme の無い URL の時、検証に失敗すること。', () => {
    // Act
    const result = recipeFormSchema.safeParse({
      ...validValues,
      inputMode: 'url',
      title: '',
      sourceUrl: 'kurashiru.com/recipes/abc',
      ingredients: [{ name: '', quantity: 0, unit: '' }],
    })

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

  it('url モードで参考 URL とサムネイルが API 形へそのまま渡ること。', () => {
    // Act
    const req = toRecipeRequest({
      ...validValues,
      inputMode: 'url',
      sourceUrl: 'https://www.kurashiru.com/recipes/abc',
      thumbnailUrl: 'https://img.example/thumb.jpg',
    })

    // Assert
    expect(req.source_url).toBe('https://www.kurashiru.com/recipes/abc')
    expect(req.thumbnail_url).toBe('https://img.example/thumb.jpg')
  })

  it('手動モードでは、残った参考 URL・サムネイルを送らないこと。', () => {
    // Act: url モードで URL を入れた後に手動モードへ切り替えたケースを想定
    const req = toRecipeRequest({
      ...validValues,
      inputMode: 'manual',
      sourceUrl: 'https://www.kurashiru.com/recipes/abc',
      thumbnailUrl: 'https://img.example/thumb.jpg',
    })

    // Assert
    expect(req.source_url).toBe('')
    expect(req.thumbnail_url).toBe('')
  })

  it('url モードでは、残った食材・調味料を送らないこと。', () => {
    // Act: 手動レシピを編集して url モードへ切り替えたケースを想定
    const req = toRecipeRequest({
      ...validValues,
      inputMode: 'url',
      sourceUrl: 'https://www.kurashiru.com/recipes/abc',
      ingredients: [{ name: '玉ねぎ', quantity: 1, unit: '個' }],
      seasoning: [{ name: '塩', quantity: 1, unit: 'g' }],
    })

    // Assert
    expect(req.cooking).toEqual([])
    expect(req.season).toEqual([])
  })

  it('タイトルが空でも参考 URL があれば、ホスト名からタイトルを補うこと。', () => {
    // Act
    const req = toRecipeRequest({
      ...validValues,
      inputMode: 'url',
      title: '',
      sourceUrl: 'https://www.kurashiru.com/recipes/abc',
    })

    // Assert
    expect(req.title).toBe('kurashiru.com のレシピ')
  })

  it('未使用の空行を含む時、送信データから除外されること。', () => {
    // Act
    const req = toRecipeRequest({
      ...validValues,
      seasoning: [{ name: '', quantity: 0, unit: '' }],
    })

    // Assert
    expect(req.season).toEqual([])
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
      source_url: '',
      thumbnail_url: '',
      owner: { id: 'u-taro', username: 'taro', avatar_url: null },
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

  it('参考 URL を持つレシピの時、初期モードが url になること。', () => {
    // Arrange
    const recipe = {
      id: 'r1',
      created_at: '2026-06-15 09:30',
      updated_at: '2026-06-15 09:30',
      cooking: [],
      season: [],
      procedure: '',
      source_url: 'https://www.kurashiru.com/recipes/abc',
      thumbnail_url: 'https://img.example/thumb.jpg',
      owner: { id: 'u-taro', username: 'taro', avatar_url: null },
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
    expect(values.inputMode).toBe('url')
  })
})
