import { apiRequest } from '@/api/client'

type HealthDTO = {
  status: string
}

export type Health = {
  status: string
}

export async function fetchHealth(): Promise<Health> {
  const dto = await apiRequest<HealthDTO>('/health')
  return { status: dto.status }
}
