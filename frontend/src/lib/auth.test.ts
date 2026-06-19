import { describe, it, expect, beforeEach } from 'vitest'
import { getToken, setToken, clearToken, isAuthed, getAuthHeaders, getUsername } from './auth'

// localStorage mock is provided by jsdom

beforeEach(() => {
  localStorage.clear()
})

describe('getToken / setToken', () => {
  it('returns null when no token', () => {
    expect(getToken()).toBeNull()
  })

  it('stores and retrieves token', () => {
    setToken('abc123')
    expect(getToken()).toBe('abc123')
  })

  it('stores token with username', () => {
    setToken('abc123', 'admin')
    expect(getToken()).toBe('abc123')
    expect(getUsername()).toBe('admin')
  })

  it('overwrites existing token', () => {
    setToken('first')
    setToken('second')
    expect(getToken()).toBe('second')
  })
})

describe('clearToken', () => {
  it('removes token and username', () => {
    setToken('abc', 'user')
    clearToken()
    expect(getToken()).toBeNull()
    expect(getUsername()).toBeNull()
  })
})

describe('isAuthed', () => {
  it('returns false when no token', () => {
    expect(isAuthed()).toBe(false)
  })

  it('returns true when token exists', () => {
    setToken('abc')
    expect(isAuthed()).toBe(true)
  })
})

describe('getAuthHeaders', () => {
  it('returns empty object when no token', () => {
    expect(getAuthHeaders()).toEqual({})
  })

  it('returns Bearer header when token exists', () => {
    setToken('mytoken')
    expect(getAuthHeaders()).toEqual({ Authorization: 'Bearer mytoken' })
  })
})

describe('getUsername', () => {
  it('returns null when no username', () => {
    expect(getUsername()).toBeNull()
  })

  it('returns stored username', () => {
    setToken('tok', 'alice')
    expect(getUsername()).toBe('alice')
  })
})
