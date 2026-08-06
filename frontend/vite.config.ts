import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const apiTarget = process.env.VITE_API_PROXY_TARGET ?? 'http://localhost:8080'
const port = Number(process.env.FRONTEND_PORT ?? 5173)
const base = normalizeBase(process.env.VITE_BASE_PATH ?? '/')
const apiPath = `${base}api`

export default defineConfig({
  base,
  plugins: [react()],
  server: {
    host: true,
    allowedHosts: ['aheron.pro'],
    port,
    proxy: {
      [apiPath]: {
        target: apiTarget,
        changeOrigin: true,
        rewrite: (path) => base === '/' ? path : path.slice(base.length - 1),
        // SSE-ответы чата не должны буферизоваться прокси
        configure: (proxy) => {
          proxy.on('proxyRes', (proxyRes) => {
            if (proxyRes.headers['content-type']?.includes('text/event-stream')) {
              proxyRes.headers['cache-control'] = 'no-cache, no-transform'
            }
          })
        },
      },
    },
    watch: {
      usePolling: true,
    },
  },
})

function normalizeBase(value: string): string {
  const path = `/${value.replace(/^\/+|\/+$/g, '')}/`
  return path === '//' ? '/' : path
}
