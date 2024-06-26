import { defineConfig } from "vite";
import Vue from "@vitejs/plugin-vue";
import VuetifyPlugin from "vite-plugin-vuetify";
import { nodePolyfills } from 'vite-plugin-node-polyfills'
import Markdown from "unplugin-vue-markdown/vite";
import * as path from "node:path";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    nodePolyfills(),
    Vue(),
    VuetifyPlugin({
      autoImport: true,
    }),
    Markdown({ markdownItOptions: {
      html: true,
      typographer: true,
    } }),
  ],
  server: {
    port: 8080,
    hmr: {
      clientPort: 80,
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
