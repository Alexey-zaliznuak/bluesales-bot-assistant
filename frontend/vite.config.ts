import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const apiTarget = process.env.VITE_API_PROXY_TARGET ?? 'http://localhost:8080'
const port = Number(process.env.FRONTEND_PORT ?? 5173)

export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    allowedHosts: ['aheron.pro'],
    port,
    proxy: {
      '/api': {
        target: apiTarget,
        changeOrigin: true,
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
