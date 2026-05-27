const TOKEN_KEY = 'apihub_token'
const USERNAME_KEY = 'apihub_username'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string, username?: string) {
  localStorage.setItem(TOKEN_KEY, token)
  if (username) {
    localStorage.setItem(USERNAME_KEY, username)
  }
}

export function getUsername(): string | null {
  return localStorage.getItem(USERNAME_KEY)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USERNAME_KEY)
}

export function isAuthed(): boolean {
  return !!getToken()
}

export function getAuthHeaders(): Record<string, string> {
  const token = getToken()
  if (!token) return {}
  return { Authorization: `Bearer ${token}` }
}
