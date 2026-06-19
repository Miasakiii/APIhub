import { describe, it, expect } from 'vitest'
import { formatUSD, formatNum, maskKey, cn } from './utils'

describe('cn', () => {
  it('merges class names', () => {
    expect(cn('foo', 'bar')).toBe('foo bar')
  })

  it('deduplicates tailwind classes', () => {
    expect(cn('px-4', 'px-8')).toBe('px-8')
  })

  it('handles conditional classes', () => {
    expect(cn('base', false && 'hidden', 'extra')).toBe('base extra')
  })

  it('handles empty input', () => {
    expect(cn()).toBe('')
  })
})

describe('formatUSD', () => {
  it('formats zero', () => {
    expect(formatUSD(0)).toBe('$0.00')
  })

  it('formats integer', () => {
    expect(formatUSD(5)).toBe('$5.00')
  })

  it('formats decimal', () => {
    expect(formatUSD(12.34)).toBe('$12.34')
  })

  it('formats large number with commas', () => {
    expect(formatUSD(1234.56)).toBe('$1,234.56')
  })

  it('formats very small amount', () => {
    expect(formatUSD(0.01)).toBe('$0.01')
  })

  it('rounds to 2 decimals', () => {
    expect(formatUSD(1.999)).toBe('$2.00')
  })
})

describe('formatNum', () => {
  it('formats zero', () => {
    expect(formatNum(0)).toBe('0')
  })

  it('formats small number', () => {
    expect(formatNum(42)).toBe('42')
  })

  it('formats thousands', () => {
    expect(formatNum(1500)).toBe('1.5K')
  })

  it('formats millions', () => {
    expect(formatNum(2500000)).toBe('2.5M')
  })

  it('formats billions', () => {
    expect(formatNum(3000000000)).toBe('3.0B')
  })

  it('formats exact 1000', () => {
    expect(formatNum(1000)).toBe('1.0K')
  })

  it('formats exact 1000000', () => {
    expect(formatNum(1000000)).toBe('1.0M')
  })
})

describe('maskKey', () => {
  it('masks long key', () => {
    const result = maskKey('sk-abcdefghijklmnop')
    expect(result).toBe('sk-a••••••••mnop')
  })

  it('masks short key with ****', () => {
    expect(maskKey('short')).toBe('****')
  })

  it('masks exactly 8 char key with ****', () => {
    expect(maskKey('12345678')).toBe('****')
  })

  it('masks 9 char key', () => {
    expect(maskKey('123456789')).toBe('1234••••••••6789')
  })

  it('preserves first 4 and last 4', () => {
    const result = maskKey('sk-xxxxxxxxxxxxxxxxxxxxxxabcdef')
    expect(result.slice(0, 4)).toBe('sk-x')
    expect(result.slice(-4)).toBe('cdef')
  })
})
