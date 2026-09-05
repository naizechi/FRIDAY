import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import {
  ApiError,
  apiRequest,
  configureAuth,
  resetAuthClient,
  setUnauthorizedHandler,
} from '@/api/client'

describe('apiRequest auth', () => {
  beforeEach(() => {
    resetAuthClient()
    configureAuth({
      getAccessToken: () => 'access-1',
      refreshTokens: async () => ({
        accessToken: 'access-2',
        refreshToken: 'refresh-2',
      }),
      updateSession: vi.fn(),
    })
    setUnauthorizedHandler(null)
  })

  afterEach(() => {
    resetAuthClient()
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('sends Authorization Bearer header', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    )
    vi.stubGlobal('fetch', fetchMock)

    await apiRequest('/me')

    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/me',
      expect.objectContaining({
        headers: expect.objectContaining({
          Authorization: 'Bearer access-1',
        }),
      }),
    )
  })

  it('refreshes once on 401 then retries', async () => {
    const updateSession = vi.fn()
    const refreshTokens = vi.fn().mockResolvedValue({
      accessToken: 'access-2',
      refreshToken: 'refresh-2',
    })

    configureAuth({
      getAccessToken: () => 'access-1',
      refreshTokens,
      updateSession,
    })

    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ message: 'unauthorized' }), {
          status: 401,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ id: 1 }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
    vi.stubGlobal('fetch', fetchMock)

    const data = await apiRequest<{ id: number }>('/me')

    expect(data).toEqual({ id: 1 })
    expect(refreshTokens).toHaveBeenCalledTimes(1)
    expect(updateSession).toHaveBeenCalledWith({
      accessToken: 'access-2',
      refreshToken: 'refresh-2',
    })
    expect(fetchMock).toHaveBeenCalledTimes(2)
  })

  it('single-flights parallel 401 refreshes', async () => {
    let resolveRefresh: (value: {
      accessToken: string
      refreshToken: string
    }) => void = () => undefined

    const refreshTokens = vi.fn(
      () =>
        new Promise<{ accessToken: string; refreshToken: string }>((resolve) => {
          resolveRefresh = resolve
        }),
    )

    configureAuth({
      getAccessToken: () => 'access-1',
      refreshTokens,
      updateSession: vi.fn(),
    })

    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ message: 'unauthorized' }), {
          status: 401,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ message: 'unauthorized' }), {
          status: 401,
          headers: { 'Content-Type': 'application/json' },
        }),
      )
      .mockImplementation(() =>
        Promise.resolve(
          new Response(JSON.stringify({ ok: true }), {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          }),
        ),
      )
    vi.stubGlobal('fetch', fetchMock)

    const p1 = apiRequest('/one')
    const p2 = apiRequest('/two')

    await vi.waitFor(() => {
      expect(refreshTokens).toHaveBeenCalledTimes(1)
    })

    resolveRefresh({ accessToken: 'access-2', refreshToken: 'refresh-2' })
    await Promise.all([p1, p2])

    expect(refreshTokens).toHaveBeenCalledTimes(1)
  })

  it('throws ApiError when refresh fails', async () => {
    configureAuth({
      getAccessToken: () => 'access-1',
      refreshTokens: async () => null,
      updateSession: vi.fn(),
    })

    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ message: 'unauthorized' }), {
          status: 401,
          headers: { 'Content-Type': 'application/json' },
        }),
      ),
    )

    await expect(apiRequest('/me')).rejects.toBeInstanceOf(ApiError)
  })
})
