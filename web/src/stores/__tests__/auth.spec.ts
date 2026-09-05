import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { useAuthStore } from '@/stores/auth'

vi.mock('@/api/auth', () => ({
  fetchMe: vi.fn(),
  login: vi.fn(),
  register: vi.fn(),
  refresh: vi.fn(),
  logout: vi.fn(),
}))

import * as authApi from '@/api/auth'

function memoryStorage(): Storage {
  const data = new Map<string, string>()
  return {
    get length() {
      return data.size
    },
    clear() {
      data.clear()
    },
    getItem(key: string) {
      return data.has(key) ? data.get(key)! : null
    },
    key(index: number) {
      return [...data.keys()][index] ?? null
    },
    removeItem(key: string) {
      data.delete(key)
    },
    setItem(key: string, value: string) {
      data.set(key, value)
    },
  }
}

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.stubGlobal('localStorage', memoryStorage())
    vi.clearAllMocks()
  })

  it('hydrates tokens and user from localStorage then fetchMe', async () => {
    localStorage.setItem('friday.access_token', 'access')
    localStorage.setItem('friday.refresh_token', 'refresh')
    localStorage.setItem(
      'friday.user',
      JSON.stringify({ id: 1, email: 'old@example.com', createdAt: '2026-01-01T00:00:00Z' }),
    )

    vi.mocked(authApi.fetchMe).mockResolvedValue({
      id: 1,
      email: 'user@example.com',
      createdAt: '2026-01-02T00:00:00Z',
    })

    const auth = useAuthStore()
    await auth.hydrate()

    expect(auth.accessToken).toBe('access')
    expect(auth.refreshToken).toBe('refresh')
    expect(auth.user?.email).toBe('user@example.com')
    expect(auth.isAuthenticated).toBe(true)
    expect(auth.ready).toBe(true)
  })

  it('persists session after login', async () => {
    vi.mocked(authApi.login).mockResolvedValue({
      user: {
        id: 2,
        email: 'new@example.com',
        createdAt: '2026-01-03T00:00:00Z',
      },
      accessToken: 'a',
      refreshToken: 'r',
      tokenType: 'bearer',
      expiresIn: 900,
    })

    const auth = useAuthStore()
    await auth.login('new@example.com', 'password123')

    expect(localStorage.getItem('friday.access_token')).toBe('a')
    expect(localStorage.getItem('friday.refresh_token')).toBe('r')
    expect(auth.user?.email).toBe('new@example.com')
  })
})
