import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import * as authApi from '@/api/auth'
import { configureAuth, setUnauthorizedHandler } from '@/api/client'
import type { Session, User } from '@/types/auth'

const ACCESS_KEY = 'friday.access_token'
const REFRESH_KEY = 'friday.refresh_token'
const USER_KEY = 'friday.user'

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string | null>(null)
  const refreshToken = ref<string | null>(null)
  const user = ref<User | null>(null)
  const ready = ref(false)

  const isAuthenticated = computed(() => Boolean(accessToken.value && user.value))

  function persist() {
    if (accessToken.value) {
      localStorage.setItem(ACCESS_KEY, accessToken.value)
    } else {
      localStorage.removeItem(ACCESS_KEY)
    }

    if (refreshToken.value) {
      localStorage.setItem(REFRESH_KEY, refreshToken.value)
    } else {
      localStorage.removeItem(REFRESH_KEY)
    }

    if (user.value) {
      localStorage.setItem(USER_KEY, JSON.stringify(user.value))
    } else {
      localStorage.removeItem(USER_KEY)
    }
  }

  function applySession(session: Session) {
    accessToken.value = session.accessToken
    refreshToken.value = session.refreshToken
    user.value = session.user
    persist()
  }

  function clearSession() {
    accessToken.value = null
    refreshToken.value = null
    user.value = null
    persist()
  }

  async function login(email: string, password: string) {
    const session = await authApi.login(email, password)
    applySession(session)
  }

  async function register(email: string, password: string) {
    const session = await authApi.register(email, password)
    applySession(session)
  }

  async function logout() {
    const currentRefresh = refreshToken.value
    clearSession()
    if (currentRefresh) {
      try {
        await authApi.logout(currentRefresh)
      } catch {
        // session already cleared locally
      }
    }
  }

  async function hydrate() {
    accessToken.value = localStorage.getItem(ACCESS_KEY)
    refreshToken.value = localStorage.getItem(REFRESH_KEY)

    const rawUser = localStorage.getItem(USER_KEY)
    if (rawUser) {
      try {
        user.value = JSON.parse(rawUser) as User
      } catch {
        user.value = null
      }
    }

    if (accessToken.value) {
      try {
        user.value = await authApi.fetchMe()
        persist()
      } catch {
        // interceptor refresh / unauthorized handler already cleared hard failures
        if (!accessToken.value) {
          user.value = null
        }
      }
    }

    ready.value = true
  }

  async function tryRefresh(): Promise<Session | null> {
    if (!refreshToken.value) {
      return null
    }

    try {
      const session = await authApi.refresh(refreshToken.value)
      applySession(session)
      return session
    } catch {
      clearSession()
      return null
    }
  }

  function wireClient() {
    configureAuth({
      getAccessToken: () => accessToken.value,
      refreshTokens: async () => {
        const session = await tryRefresh()
        if (!session) {
          return null
        }
        return {
          accessToken: session.accessToken,
          refreshToken: session.refreshToken,
        }
      },
      updateSession: (tokens) => {
        accessToken.value = tokens.accessToken
        refreshToken.value = tokens.refreshToken
        persist()
      },
    })

    setUnauthorizedHandler(() => {
      clearSession()
    })
  }

  return {
    accessToken,
    refreshToken,
    user,
    ready,
    isAuthenticated,
    login,
    register,
    logout,
    hydrate,
    clearSession,
    applySession,
    wireClient,
    tryRefresh,
  }
})
