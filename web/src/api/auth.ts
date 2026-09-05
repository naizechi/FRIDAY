import { apiRequest } from '@/api/client'
import type { Session, User } from '@/types/auth'

type UserDTO = {
  id: number
  email: string
  created_at: string
}

type SessionDTO = {
  user: UserDTO
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
}

function mapUser(dto: UserDTO): User {
  return {
    id: dto.id,
    email: dto.email,
    createdAt: dto.created_at,
  }
}

function mapSession(dto: SessionDTO): Session {
  return {
    user: mapUser(dto.user),
    accessToken: dto.access_token,
    refreshToken: dto.refresh_token,
    tokenType: dto.token_type,
    expiresIn: dto.expires_in,
  }
}

export async function register(email: string, password: string): Promise<Session> {
  const dto = await apiRequest<SessionDTO>('/auth/register', {
    method: 'POST',
    body: { email, password },
  })
  return mapSession(dto)
}

export async function login(email: string, password: string): Promise<Session> {
  const dto = await apiRequest<SessionDTO>('/auth/login', {
    method: 'POST',
    body: { email, password },
  })
  return mapSession(dto)
}

export async function refresh(refreshToken: string): Promise<Session> {
  const dto = await apiRequest<SessionDTO>('/auth/refresh', {
    method: 'POST',
    body: { refresh_token: refreshToken },
    skipAuth: true,
    skipRefresh: true,
  })
  return mapSession(dto)
}

export async function logout(refreshToken: string): Promise<void> {
  await apiRequest<void>('/auth/logout', {
    method: 'POST',
    body: { refresh_token: refreshToken },
    skipAuth: true,
    skipRefresh: true,
  })
}

export async function fetchMe(): Promise<User> {
  const dto = await apiRequest<UserDTO>('/me')
  return mapUser(dto)
}
