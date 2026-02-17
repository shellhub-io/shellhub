import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import * as path from "node:path";

export default defineConfig({
  plugins: [react()],
  base: "/v2/",
  server: {
    port: 8080,
    allowedHosts: true,
    proxy: {
      "/v2/docs": {
        target: "http://localhost:8083",
        ws: true,
      },
      "/v2/blog": {
        target: "http://localhost:8084",
        ws: true,
      },
      "/v2/ui": {
        target: "http://localhost:8082",
        ws: true,
      },
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
  define: {
    "process.env": process.env,
  },
});
