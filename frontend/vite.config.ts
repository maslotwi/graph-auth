import path from "path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig, loadEnv } from "vite"

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "")
  const useMocks = env.VITE_USE_MOCKS === "true"

  return {
    plugins: [react(), tailwindcss()],
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
      },
    },
    server: {
      proxy: useMocks
        ? undefined
        : {
            "/api": {
              target: "http://localhost:8080",
              changeOrigin: true,
            },
          },
    },
  }
})
