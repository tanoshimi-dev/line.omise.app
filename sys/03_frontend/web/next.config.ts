import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  // `standalone` keeps the production Docker image small (see Dockerfile).
  output: 'standalone',
}

export default nextConfig
