/// <reference types="vitest/config" />
import { defineConfig, type Plugin } from "vite";
import react, { reactCompilerPreset } from "@vitejs/plugin-react";
import babel from "@rolldown/plugin-babel";
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
    react(),
    babel({ presets: [reactCompilerPreset()] }),
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
      "@": path.resolve(import.meta.dirname, "src"),
    },
  },
  build: {
    rolldownOptions: {
      input: {
        main: path.resolve(import.meta.dirname, "index.html"),
        migrate: path.resolve(import.meta.dirname, "migrate.html"),
      },
    },
  },
  test: {
    environment: "jsdom",
    setupFiles: [path.resolve(import.meta.dirname, "./src/tests/setup.ts")],
    include: ["src/**/*.test.{ts,tsx}"],
  },
});
