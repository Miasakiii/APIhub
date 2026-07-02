import { describe, it, expect } from 'vitest'
import { navMain, navBottom, allNav } from './nav'

describe('navMain', () => {
  it('has 3 items', () => {
    expect(navMain).toHaveLength(3)
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

describe('navBottom', () => {
  it('has settings', () => {
    expect(navBottom).toHaveLength(1)
    expect(navBottom[0].id).toBe('settings')
  })
})

describe('allNav', () => {
  it('is the sum of all nav arrays', () => {
    expect(allNav).toHaveLength(navMain.length + navBottom.length)
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