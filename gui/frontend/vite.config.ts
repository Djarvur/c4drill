import { defineConfig } from "vite";

// base './' so the built assets resolve through the Wails asset server and
// the plain HTTP fallback alike (both serve the dist root directly).
export default defineConfig({
  base: "./",
  server: {
    port: 5279,
    // Dev convenience: proxy the backend API to `go run ./gui --serve`.
    proxy: {
      "/api": {
        target: "http://127.0.0.1:5278",
        changeOrigin: false,
      },
    },
  },
  build: {
    outDir: "dist",
    sourcemap: false,
    target: "es2022",
  },
});
