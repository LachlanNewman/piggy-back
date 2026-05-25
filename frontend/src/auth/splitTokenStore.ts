const COOKIE_NAME = 'oidc_rt'

export class SplitTokenStore {
  #map = new Map<string, string>()

  async set(key: string, value: string): Promise<void> {
    const data = JSON.parse(value) as Record<string, unknown>
    const { refresh_token, ...rest } = data
    this.#map.set(key, JSON.stringify(rest))
    if (refresh_token && typeof refresh_token === 'string') {
      writeCookie(refresh_token)
    }
  }

  async get(key: string): Promise<string | null> {
    const stored = this.#map.get(key)
    if (stored) {
      const data = JSON.parse(stored) as Record<string, unknown>
      const rt = readCookie()
      if (rt) data.refresh_token = rt
      return JSON.stringify(data)
    }
    // After a page reload the Map is empty. If we have a refresh token cookie,
    // return a minimal expired stub so signinSilent() uses the refresh_token grant.
    const rt = readCookie()
    if (!rt) return null
    return JSON.stringify({
      access_token: '',
      token_type: 'Bearer',
      profile: {},
      expires_at: 0,
      refresh_token: rt,
    })
  }

  async remove(key: string): Promise<string | null> {
    const existing = await this.get(key)
    this.#map.delete(key)
    clearCookie()
    return existing
  }

  async getAllKeys(): Promise<string[]> {
    return Array.from(this.#map.keys())
  }
}

function writeCookie(value: string): void {
  const secure = location.protocol === 'https:' ? '; Secure' : ''
  document.cookie = `${COOKIE_NAME}=${encodeURIComponent(value)}; path=/; SameSite=Strict${secure}`
}

function readCookie(): string | null {
  const match = document.cookie.split('; ').find(r => r.startsWith(`${COOKIE_NAME}=`))
  return match ? decodeURIComponent(match.split('=').slice(1).join('=')) : null
}

function clearCookie(): void {
  document.cookie = `${COOKIE_NAME}=; path=/; max-age=0; SameSite=Strict`
}
