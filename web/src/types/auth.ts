export type User = {
  id: number
  email: string
  createdAt: string
}

export type Session = {
  user: User
  accessToken: string
  refreshToken: string
  tokenType: string
  expiresIn: number
}
