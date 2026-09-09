import { z } from 'zod'

export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
    this.name = 'ApiError'
  }
}

// --- Schemas ---

const UserMeSchema = z.object({
  id: z.number(),
  auth_subject: z.string(),
  first_name: z.string(),
  last_name: z.string(),
  email: z.string(),
  profile_complete: z.boolean(),
})

const NearbyUserSchema = z.object({
  id: z.number(),
  auth_subject: z.string(),
  first_name: z.string(),
  last_name: z.string(),
})

const RideRequestSchema = z.object({
  id: z.string(),
  status: z.string(),
  pickup_address: z.string(),
  dropoff_address: z.string(),
  requested_at: z.string(),
  expires_at: z.string(),
})

const IncomingRequestSchema = z.object({
  id: z.string(),
  rider_id: z.string(),
  rider_first_name: z.string(),
  rider_last_name: z.string(),
  pickup_address: z.string(),
  dropoff_address: z.string(),
  requested_at: z.string(),
  expires_at: z.string(),
})

const ErrorResponseSchema = z.object({ error: z.string().optional() })

// --- Exported types (inferred from schemas) ---

export type UserMe = z.infer<typeof UserMeSchema>
export type NearbyUser = z.infer<typeof NearbyUserSchema>
export type RideRequest = z.infer<typeof RideRequestSchema>
export type IncomingRequest = z.infer<typeof IncomingRequestSchema>

// --- Input param interfaces (not API responses, no schemas needed) ---

export interface CreateUserParams {
  first_name: string
  last_name: string
  email: string
  date_of_birth: string
  weight: number
  gender: string
}

export interface CreateRideRequestParams {
  pickupAddress: string
  dropoffAddress: string
  driverID: string
}

// --- Base URL ---

// Empty in dev, where Vite proxies /api to the backend (see vite.config.ts).
// Set VITE_API_BASE_URL when the frontend is served separately from the API,
// e.g. a Render static site calling the backend web service. Baked in at build
// time, so changing it requires a rebuild.
const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/$/, '')

function apiUrl(path: string): string {
  return `${API_BASE_URL}${path}`
}

// --- Auth ---

type TokenProvider = () => string | undefined

let getToken: TokenProvider = () => undefined

// App registers the provider once. The token is read at call time rather than
// captured, so requests issued after a silent renew use the fresh one.
export function setTokenProvider(fn: TokenProvider): void {
  getToken = fn
}

function authHeaders(extra?: Record<string, string>): Record<string, string> {
  const token = getToken()
  return { ...extra, ...(token ? { Authorization: `Bearer ${token}` } : {}) }
}

// --- Client ---

class BackendClient {
  async getUserMe(): Promise<UserMe | null> {
    const res = await fetch(apiUrl('/api/v1/users/me'), { headers: authHeaders() })
    if (res.status === 404) return null
    if (!res.ok) throw new ApiError(res.status, 'profile check failed')
    return UserMeSchema.parse(await res.json())
  }

  async createUser(body: CreateUserParams): Promise<{ id: number }> {
    const res = await fetch(apiUrl('/api/v1/users'), {
      method: 'POST',
      headers: authHeaders({ 'Content-Type': 'application/json' }),
      body: JSON.stringify(body),
    })
    const data = ErrorResponseSchema.passthrough().parse(await res.json())
    if (!res.ok) throw new ApiError(res.status, data.error ?? 'could not create user')
    return data as { id: number }
  }

  async pushLocation(lat: number, lng: number): Promise<void> {
    await fetch(apiUrl('/api/v1/location'), {
      method: 'POST',
      headers: authHeaders({ 'Content-Type': 'application/json' }),
      body: JSON.stringify({ lat, lng }),
    })
  }

  async getNearbyUsers(): Promise<NearbyUser[]> {
    const res = await fetch(apiUrl('/api/v1/users/nearby'), { headers: authHeaders() })
    if (!res.ok) throw new ApiError(res.status, 'could not fetch nearby users')
    return z.array(NearbyUserSchema).parse(await res.json())
  }

  async createRideRequest(params: CreateRideRequestParams): Promise<{ id: string }> {
    const res = await fetch(apiUrl('/api/v1/ride-requests'), {
      method: 'POST',
      headers: authHeaders({ 'Content-Type': 'application/json' }),
      body: JSON.stringify({
        pickup_address: params.pickupAddress,
        dropoff_address: params.dropoffAddress,
        driver_id: params.driverID,
      }),
    })
    const data = ErrorResponseSchema.passthrough().parse(await res.json())
    if (!res.ok) throw new ApiError(res.status, data.error ?? 'could not create ride request')
    return data as { id: string }
  }

  async getRideRequest(id: string): Promise<RideRequest> {
    const res = await fetch(apiUrl(`/api/v1/ride-requests/${id}`), { headers: authHeaders() })
    if (!res.ok) throw new ApiError(res.status, 'could not get ride request')
    return RideRequestSchema.parse(await res.json())
  }

  async getIncomingRequests(): Promise<IncomingRequest[]> {
    const res = await fetch(apiUrl('/api/v1/ride-requests/incoming'), { headers: authHeaders() })
    if (!res.ok) return []
    return z.array(IncomingRequestSchema).parse(await res.json())
  }

  async acceptRideRequest(id: string): Promise<void> {
    await fetch(apiUrl(`/api/v1/ride-requests/${id}/accept`), { method: 'PATCH', headers: authHeaders() })
  }

  async declineRideRequest(id: string): Promise<void> {
    await fetch(apiUrl(`/api/v1/ride-requests/${id}/decline`), { method: 'PATCH', headers: authHeaders() })
  }
}

export const backendClient = new BackendClient()
