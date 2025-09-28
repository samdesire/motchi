import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const API_URI = process.env.VITE_LOG_LEVEL === 'production'
  ? 'http://example.com'    // in prod
  : 'http://localhost:5815' // in dev

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      },
      '/websocket': {
        target: 'ws://localhost:8080',
        ws: true,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/websocket/, '')
      }
    }
  }
})
