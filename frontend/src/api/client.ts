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
  auth_subject: string
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

// --- Client ---

class BackendClient {
  async getUserMe(sub: string): Promise<UserMe | null> {
    const res = await fetch(`/api/v1/users/me?sub=${encodeURIComponent(sub)}`)
    if (res.status === 404) return null
    if (!res.ok) throw new ApiError(res.status, 'profile check failed')
    return UserMeSchema.parse(await res.json())
  }

  async createUser(body: CreateUserParams): Promise<{ id: number }> {
    const res = await fetch('/api/v1/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    const data = ErrorResponseSchema.passthrough().parse(await res.json())
    if (!res.ok) throw new ApiError(res.status, data.error ?? 'could not create user')
    return data as { id: number }
  }

  async pushLocation(sub: string, lat: number, lng: number): Promise<void> {
    await fetch(`/api/v1/location?sub=${encodeURIComponent(sub)}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ lat, lng }),
    })
  }

  async getNearbyUsers(sub: string): Promise<NearbyUser[]> {
    const res = await fetch(`/api/v1/users/nearby?sub=${encodeURIComponent(sub)}`)
    if (!res.ok) throw new ApiError(res.status, 'could not fetch nearby users')
    return z.array(NearbyUserSchema).parse(await res.json())
  }

  async createRideRequest(sub: string, params: CreateRideRequestParams): Promise<{ id: string }> {
    const res = await fetch(`/api/v1/ride-requests?sub=${encodeURIComponent(sub)}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
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
    const res = await fetch(`/api/v1/ride-requests/${id}`)
    if (!res.ok) throw new ApiError(res.status, 'could not get ride request')
    return RideRequestSchema.parse(await res.json())
  }

  async getIncomingRequests(sub: string): Promise<IncomingRequest[]> {
    const res = await fetch(`/api/v1/ride-requests/incoming?sub=${encodeURIComponent(sub)}`)
    if (!res.ok) return []
    return z.array(IncomingRequestSchema).parse(await res.json())
  }

  async acceptRideRequest(id: string, sub: string): Promise<void> {
    await fetch(`/api/v1/ride-requests/${id}/accept?sub=${encodeURIComponent(sub)}`, { method: 'PATCH' })
  }

  async declineRideRequest(id: string, sub: string): Promise<void> {
    await fetch(`/api/v1/ride-requests/${id}/decline?sub=${encodeURIComponent(sub)}`, { method: 'PATCH' })
  }
}

export const backendClient = new BackendClient()
