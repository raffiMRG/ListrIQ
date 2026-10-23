import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

const backend = 'http://localhost:8080'

// Saat `npm run dev`, /api dan /ws diteruskan ke backend Go.
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': backend,
      '/ws': {
        target: backend,
        ws: true,
        // backend menolak WebSocket lintas origin; samakan origin agar lolos saat dev
        configure: (proxy) => proxy.on('proxyReqWs', (req) => req.setHeader('origin', backend)),
      },
    },
  },
})
