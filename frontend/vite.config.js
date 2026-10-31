import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  // Lee el .env de la raíz del repo (solo expone variables VITE_*)
  envDir: '..',
  server: {
    proxy: { '/api': 'http://localhost:8080' },
  },
})
