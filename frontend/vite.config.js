import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig, loadEnv } from 'vite'

export default defineConfig(({ mode }) => {
  // Lee el .env de la raíz del repo (al navegador solo llegan las variables VITE_*)
  const env = loadEnv(mode, '..', '')
  return {
    plugins: [react(), tailwindcss()],
    envDir: '..',
    server: {
      proxy: { '/api': `http://localhost:${env.APP_PORT || 8080}` },
    },
  }
})
