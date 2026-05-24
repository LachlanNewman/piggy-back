export class ApiError extends Error {
  constructor(status, message) {
    super(message)
    this.status = status
    this.name = 'ApiError'
  }
}

class BackendClient {
  async getUserMe(sub) {
    const res = await fetch(`/api/v1/users/me?sub=${encodeURIComponent(sub)}`)
    if (res.status === 404) return null
    if (!res.ok) throw new ApiError(res.status, 'profile check failed')
    return res.json()
  }

  async createUser(body) {
    const res = await fetch('/api/v1/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    const data = await res.json()
    if (!res.ok) throw new ApiError(res.status, data.error ?? 'could not create user')
    return data
  }

  async pushLocation(sub, lat, lng) {
    await fetch(`/api/v1/location?sub=${encodeURIComponent(sub)}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ lat, lng }),
    })
  }

  async getNearbyUsers(sub) {
    const res = await fetch(`/api/v1/users/nearby?sub=${encodeURIComponent(sub)}`)
    if (!res.ok) throw new ApiError(res.status, 'could not fetch nearby users')
    return res.json()
  }

  async createRideRequest(sub, { pickupAddress, dropoffAddress, driverID }) {
    const res = await fetch(`/api/v1/ride-requests?sub=${encodeURIComponent(sub)}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ pickup_address: pickupAddress, dropoff_address: dropoffAddress, driver_id: driverID }),
    })
    const data = await res.json()
    if (!res.ok) throw new ApiError(res.status, data.error ?? 'could not create ride request')
    return data
  }

  async getRideRequest(id) {
    const res = await fetch(`/api/v1/ride-requests/${id}`)
    if (!res.ok) throw new ApiError(res.status, 'could not get ride request')
    return res.json()
  }

  async getIncomingRequests(sub) {
    const res = await fetch(`/api/v1/ride-requests/incoming?sub=${encodeURIComponent(sub)}`)
    if (!res.ok) return []
    return res.json()
  }

  async acceptRideRequest(id, sub) {
    await fetch(`/api/v1/ride-requests/${id}/accept?sub=${encodeURIComponent(sub)}`, { method: 'PATCH' })
  }

  async declineRideRequest(id, sub) {
    await fetch(`/api/v1/ride-requests/${id}/decline?sub=${encodeURIComponent(sub)}`, { method: 'PATCH' })
  }
}

export const backendClient = new BackendClient()
