import { describe, it, expect } from 'vitest'
import { navMain, navMore, navBottom, allNav } from './nav'

describe('navMain', () => {
  it('has 4 items', () => {
    expect(navMain).toHaveLength(4)
  })

  it('starts with dashboard', () => {
    expect(navMain[0].id).toBe('dashboard')
    expect(navMain[0].path).toBe('/')
  })

  it('all items have required fields', () => {
    for (const item of navMain) {
      expect(item.id).toBeTruthy()
      expect(item.label).toBeTruthy()
      expect(item.icon).toBeDefined()
      expect(item.path).toBeTruthy()
    }
  })
})

describe('navMore', () => {
  it('has 6 items', () => {
    expect(navMore).toHaveLength(6)
  })

  it('includes playground', () => {
    expect(navMore.some(i => i.id === 'playground')).toBe(true)
  })
})

describe('navBottom', () => {
  it('has settings', () => {
    expect(navBottom).toHaveLength(1)
    expect(navBottom[0].id).toBe('settings')
  })
})

describe('allNav', () => {
  it('is the sum of all nav arrays', () => {
    expect(allNav).toHaveLength(navMain.length + navMore.length + navBottom.length)
  })

  it('has unique ids', () => {
    const ids = allNav.map(i => i.id)
    expect(new Set(ids).size).toBe(ids.length)
  })

  it('has unique paths', () => {
    const paths = allNav.map(i => i.path)
    expect(new Set(paths).size).toBe(paths.length)
  })
})
