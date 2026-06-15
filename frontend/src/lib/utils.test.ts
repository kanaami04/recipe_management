import { describe, expect, it } from 'vitest'

import { cn } from './utils'

describe('cn', () => {
  it('複数のクラス名を渡した時、空白区切りで結合されること。', () => {
    // Arrange
    const base = 'px-2'
    const extra = 'text-sm'

    // Act
    const result = cn(base, extra)

    // Assert
    expect(result).toBe('px-2 text-sm')
  })

  it('競合する Tailwind クラスを渡した時、後勝ちでマージされること。', () => {
    // Arrange
    const first = 'px-2'
    const second = 'px-4'

    // Act
    const result = cn(first, second)

    // Assert
    expect(result).toBe('px-4')
  })

  it('falsy な値を渡した時、無視されること。', () => {
    // Arrange
    const condition = false

    // Act
    const result = cn('px-2', condition && 'hidden')

    // Assert
    expect(result).toBe('px-2')
  })
})
