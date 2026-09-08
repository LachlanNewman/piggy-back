import type { NextConfig } from 'next'

const backendUrl = process.env.BACKEND_URL ?? 'http://localhost:8080'

const nextConfig: NextConfig = {
  reactStrictMode: true,
  // The repo keeps its own root CLAUDE.md; don't auto-generate per-app agent docs.
  agentRules: false,
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${backendUrl}/api/:path*`,
      },
    ]
  },
}

export default nextConfig
