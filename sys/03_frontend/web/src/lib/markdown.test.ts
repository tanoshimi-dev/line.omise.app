import { describe, expect, it } from 'vitest'
import { excerpt } from './markdown'

describe('excerpt', () => {
  it('strips headings, emphasis and links down to plain text', () => {
    const markdown = '## Heading\n\nSome **bold** text with a [link](https://example.com).'
    expect(excerpt(markdown, 200)).toBe('Heading Some bold text with a link.')
  })

  it('strips code blocks and inline code entirely', () => {
    const markdown = 'Before\n\n```\nconst x = 1\n```\n\nAfter `inline` code.'
    const result = excerpt(markdown, 200)
    expect(result).not.toContain('const x')
    expect(result).not.toContain('`')
  })

  it('truncates and appends an ellipsis when longer than the limit', () => {
    const markdown = 'a'.repeat(150)
    const result = excerpt(markdown, 100)
    expect(result).toHaveLength(101) // 100 chars + ellipsis
    expect(result.endsWith('…')).toBe(true)
  })

  it('does not truncate text shorter than the limit', () => {
    const result = excerpt('short text', 100)
    expect(result).toBe('short text')
  })
})
