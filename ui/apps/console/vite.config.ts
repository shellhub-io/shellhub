/// <reference types="vitest/config" />
import { defineConfig, type Plugin } from "vite";
import react from "@vitejs/plugin-react";
import { nodePolyfills } from "vite-plugin-node-polyfills";
import * as path from "node:path";

function healthcheck(): Plugin {
  return {
    name: "healthcheck",
    configureServer(server) {
      server.middlewares.use("/health", (_req, res) => {
        res.setHeader("Content-Type", "text/plain");
        res.end("OK");
      });
    },
  };
}

export default defineConfig({
  plugins: [
    react({
      babel: {
        plugins: ["babel-plugin-react-compiler"],
      },
    }),
    nodePolyfills({
      include: ["buffer", "crypto", "stream"],
    }),
    healthcheck(),
  ],
  base: "/",
  server: {
    port: 8080,
    allowedHosts: true,
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
  build: {
    rollupOptions: {
      input: {
        main: path.resolve(__dirname, "index.html"),
        migrate: path.resolve(__dirname, "migrate.html"),
      },
    },
  },
  test: {
    environment: "jsdom",
    setupFiles: [path.resolve(__dirname, "./src/tests/setup.ts")],
  },
});
