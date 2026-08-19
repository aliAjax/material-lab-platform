import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 15173,
    proxy: {
      "/api": "http://127.0.0.1:18080",
      "/health": "http://127.0.0.1:18080",
    },
  },
  test: { environment: "jsdom", setupFiles: "./src/test/setup.ts" },
});
