import { describe, expect, it } from 'vitest'

import { renderWithProviders, screen } from './renderWithProviders'

describe('renderWithProviders', () => {
  it('要素を描画した時、Provider 配下で DOM に出力されること。', () => {
    // Arrange
    const element = <p>こんにちは</p>

    // Act
    renderWithProviders(element)

    // Assert
    expect(screen.getByText('こんにちは')).toBeInTheDocument()
  })
})
