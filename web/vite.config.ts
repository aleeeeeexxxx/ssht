import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:6001',
        changeOrigin: true,
        ws: true,
      },
    },
  },
  build: {
    outDir: '../cmd/ssht/dist',
    emptyOutDir: true,
  },
})
