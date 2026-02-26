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
  plugins: [
    react(),
    nodePolyfills({
      // buffer, crypto, stream are needed by node-rsa and sshpk for SSH key parsing/signing.
      // Note: vault-crypto.ts uses browser-native crypto.subtle (Web Crypto API),
      // which is NOT affected by this polyfill — it only intercepts import/require('crypto').
      include: ["buffer", "crypto", "stream"],
    }),
    healthcheck(),
  ],
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
    setupFiles: [path.resolve(__dirname, "./src/test/setup.ts")],
  },
});
