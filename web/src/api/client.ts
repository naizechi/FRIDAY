export class ApiError extends Error {
  readonly status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

type RequestOptions = Omit<RequestInit, 'body'> & {
  body?: unknown
  skipAuth?: boolean
  skipRefresh?: boolean
}

type UnauthorizedHandler = () => void
type TokenGetter = () => string | null
type TokenPair = { accessToken: string; refreshToken: string }
type TokenRefresher = () => Promise<TokenPair | null>
type SessionUpdater = (tokens: TokenPair) => void

let onUnauthorized: UnauthorizedHandler | null = null
let getAccessToken: TokenGetter | null = null
let refreshTokens: TokenRefresher | null = null
let updateSession: SessionUpdater | null = null

let refreshPromise: Promise<TokenPair | null> | null = null

export function setUnauthorizedHandler(handler: UnauthorizedHandler | null) {
  onUnauthorized = handler
}

export function configureAuth(options: {
  getAccessToken: TokenGetter
  refreshTokens: TokenRefresher
  updateSession: SessionUpdater
}) {
  getAccessToken = options.getAccessToken
  refreshTokens = options.refreshTokens
  updateSession = options.updateSession
}

/** Test helper: clear in-flight refresh and auth hooks. */
export function resetAuthClient() {
  onUnauthorized = null
  getAccessToken = null
  refreshTokens = null
  updateSession = null
  refreshPromise = null
}

export async function apiRequestWithHeaders<T>(
  path: string,
  options: RequestOptions = {},
): Promise<{ data: T; headers: Headers }> {
  const response = await performRequest(path, options)

  if (response.status === 204) {
    return { data: undefined as T, headers: response.headers }
  }

  const contentType = response.headers.get('content-type') ?? ''
  if (!contentType.includes('application/json')) {
    return { data: undefined as T, headers: response.headers }
  }

  return {
    data: (await response.json()) as T,
    headers: response.headers,
  }
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { data } = await apiRequestWithHeaders<T>(path, options)
  return data
}

async function performRequest(path: string, options: RequestOptions = {}): Promise<Response> {
  const response = await rawFetch(path, options)

  if (response.status !== 401 || options.skipRefresh) {
    return settleResponse(response)
  }

  const tokens = await singleFlightRefresh()
  if (!tokens) {
    onUnauthorized?.()
    return settleResponse(response)
  }

  updateSession?.(tokens)
  const retry = await rawFetch(path, { ...options, skipRefresh: true })
  return settleResponse(retry)
}

async function singleFlightRefresh(): Promise<TokenPair | null> {
  if (!refreshTokens) {
    return null
  }

  if (!refreshPromise) {
    refreshPromise = refreshTokens().finally(() => {
      refreshPromise = null
    })
  }

  return refreshPromise
}

async function rawFetch(path: string, options: RequestOptions = {}): Promise<Response> {
  const { body, headers, skipAuth, skipRefresh: _skipRefresh, ...rest } = options

  const requestHeaders: Record<string, string> = {
    Accept: 'application/json',
    ...(body !== undefined ? { 'Content-Type': 'application/json' } : {}),
    ...(headers as Record<string, string> | undefined),
  }

  if (!skipAuth) {
    const token = getAccessToken?.()
    if (token) {
      requestHeaders.Authorization = `Bearer ${token}`
    }
  }

  return fetch(`/api/v1${path}`, {
    ...rest,
    headers: requestHeaders,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })
}

async function settleResponse(response: Response): Promise<Response> {
  if (response.status === 401) {
    onUnauthorized?.()
  }

  if (!response.ok) {
    let message = `Request failed (${response.status})`
    try {
      const payload = (await response.json()) as { message?: string }
      if (payload.message) {
        message = payload.message
      }
    } catch {
      // keep default message
    }
    throw new ApiError(message, response.status)
  }

  return response
}
