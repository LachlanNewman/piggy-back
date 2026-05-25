export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
    this.name = 'ApiError'
  }
}

export interface UserMe {
  id: number
  auth_subject: string
  first_name: string
  last_name: string
  email: string
  profile_complete: boolean
}

export interface NearbyUser {
  id: number
  auth_subject: string
  first_name: string
  last_name: string
}

export interface RideRequest {
  id: string
  status: string
  pickup_address: string
  dropoff_address: string
  requested_at: string
  expires_at: string
}

export interface IncomingRequest {
  id: string
  rider_id: string
  rider_first_name: string
  rider_last_name: string
  pickup_address: string
  dropoff_address: string
  requested_at: string
  expires_at: string
}

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

class BackendClient {
  async getUserMe(sub: string): Promise<UserMe | null> {
    const res = await fetch(`/api/v1/users/me?sub=${encodeURIComponent(sub)}`)
    if (res.status === 404) return null
    if (!res.ok) throw new ApiError(res.status, 'profile check failed')
    return res.json() as Promise<UserMe>
  }

  async createUser(body: CreateUserParams): Promise<{ id: number }> {
    const res = await fetch('/api/v1/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    const data = await res.json() as { id: number; error?: string }
    if (!res.ok) throw new ApiError(res.status, data.error ?? 'could not create user')
    return data
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
    return res.json() as Promise<NearbyUser[]>
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
    const data = await res.json() as { id: string; error?: string }
    if (!res.ok) throw new ApiError(res.status, data.error ?? 'could not create ride request')
    return data
  }

  async getRideRequest(id: string): Promise<RideRequest> {
    const res = await fetch(`/api/v1/ride-requests/${id}`)
    if (!res.ok) throw new ApiError(res.status, 'could not get ride request')
    return res.json() as Promise<RideRequest>
  }

  async getIncomingRequests(sub: string): Promise<IncomingRequest[]> {
    const res = await fetch(`/api/v1/ride-requests/incoming?sub=${encodeURIComponent(sub)}`)
    if (!res.ok) return []
    return res.json() as Promise<IncomingRequest[]>
  }

  async acceptRideRequest(id: string, sub: string): Promise<void> {
    await fetch(`/api/v1/ride-requests/${id}/accept?sub=${encodeURIComponent(sub)}`, { method: 'PATCH' })
  }

  async declineRideRequest(id: string, sub: string): Promise<void> {
    await fetch(`/api/v1/ride-requests/${id}/decline?sub=${encodeURIComponent(sub)}`, { method: 'PATCH' })
  }
}

export const backendClient = new BackendClient()
