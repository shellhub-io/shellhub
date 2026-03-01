/// <reference types="vitest/config" />
import { defineConfig, type Plugin } from "vite";
import react from "@vitejs/plugin-react";
import { nodePolyfills } from "vite-plugin-node-polyfills";
import * as path from "node:path";

function healthcheck(): Plugin {
  return {
    name: "healthcheck",
    configureServer(server) {
      server.middlewares.use("/healthcheck", (_req, res) => {
        res.setHeader("Content-Type", "text/plain");
        res.end("OK");
      });
    },
  };
}

export default defineConfig({
  plugins: [react(), nodePolyfills(), healthcheck()],
  base: "/v2/ui/",
  server: {
    port: 8080,
    allowedHosts: true,
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
  },
});
