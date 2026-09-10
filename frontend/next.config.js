/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // 'standalone' facilita construir una imagen Docker de producción
  // pequeña en fases posteriores (Fase 14).
  output: 'standalone',
};

module.exports = nextConfig;
