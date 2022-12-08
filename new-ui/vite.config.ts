import { defineConfig } from "vitest/config";
import vue from "@vitejs/plugin-vue";
// https://github.com/vuetifyjs/vuetify-loader/tree/next/packages/vite-plugin
// @ts-ignore
import vuetify from "vite-plugin-vuetify";
import inject from "@rollup/plugin-inject";
import NodeGlobalsPolyfillPlugin from "@esbuild-plugins/node-globals-polyfill";
import polyfillNode from "rollup-plugin-polyfill-node";
import { splitVendorChunkPlugin } from "vite";
import { fileURLToPath, URL } from "url";

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    port: 8080,
    hmr: {
      clientPort: 80,
    },
  },
  plugins: [
    splitVendorChunkPlugin(),
    vue({
      template: {
        compilerOptions: {
          isCustomElement: (tag) => tag.includes("v-list-item-group") || tag.includes("font-awesome-icon"),
        },
      },
    }),
    vuetify({ autoImport: true }),
    NodeGlobalsPolyfillPlugin({
      process: true,
      buffer: true,
    }),
  ],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
    
  },
  define: { 
    "process.env": process.env,
  },
  test: {
    // environment: "jsdom",
    environment: "happy-dom",
    globals: true,
    setupFiles: "vuetify.config.ts",
    deps: {
      inline: ["vuetify"],
    },
    update: true,
  },
  build: {
    rollupOptions: {
      plugins: [
        // @ts-ignore
        polyfillNode(),
        NodeGlobalsPolyfillPlugin({
          process: true,
          buffer: true,
        }),
      ],
    },
  },
});
