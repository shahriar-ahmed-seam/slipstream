/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  experimental: {
    typedRoutes: false,
  },
  env: {
    NEXT_PUBLIC_API_BASE: process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8080",
    NEXT_PUBLIC_WS_BASE: process.env.NEXT_PUBLIC_WS_BASE || "ws://localhost:8080",
  },
};

export default nextConfig;