import { defineConfig } from "vitest/config";
import vue from "@vitejs/plugin-vue";
// https://github.com/vuetifyjs/vuetify-loader/tree/next/packages/vite-plugin
import vuetify from "vite-plugin-vuetify";
import NodeGlobalsPolyfillPlugin from "@esbuild-plugins/node-globals-polyfill";
import polyfillNode from "rollup-plugin-polyfill-node";
import { splitVendorChunkPlugin } from "vite";
import { fileURLToPath, URL } from "url";
import Markdown from "vite-plugin-vue-markdown";

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
      include: [/\.vue$/, /\.md$/],
    }),
    Markdown({
      markdownItOptions: {
        html: true,
        typographer: true,
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
  build: {
    rollupOptions: {
      plugins: [
        polyfillNode(),
        NodeGlobalsPolyfillPlugin({
          process: true,
          buffer: true,
        }),
      ],
    },
  },
});
