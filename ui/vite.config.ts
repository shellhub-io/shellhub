import { defineConfig } from "vite";
import Vue from "@vitejs/plugin-vue";
import VuetifyPlugin from "vite-plugin-vuetify";
import { nodePolyfills } from 'vite-plugin-node-polyfills'
import Markdown from "unplugin-vue-markdown/vite";
import { fileURLToPath } from "node:url";
import * as path from "node:path";

const dirname = path.dirname(fileURLToPath(import.meta.url));

// https://vitejs.dev/config/
export default defineConfig({
  build: {
    rollupOptions: {
      input: {
        main: path.resolve(dirname, "index.html"),
        admin: path.resolve(dirname, "admin/index.html"),
      },
    },
  },
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
  css: {
    preprocessorOptions: {
      scss: {
        api: "modern",
      },
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
      "@ui": path.resolve(__dirname, "src"),
      "@admin": path.resolve(__dirname, "admin/src"),
    },
  },
  define: {
    "process.env": process.env,
  },
});
