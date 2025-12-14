import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  base: '/darts/',
  server: {
    proxy: {
      // Proxy API calls to backend during development
      '/darts/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
