import { defineConfig } from "vite";
import Vue from "@vitejs/plugin-vue";
import vuetify from "vite-plugin-vuetify";
import Markdown from "unplugin-vue-markdown/vite";
import * as path from "node:path";

function polyfillNode() {
  return {
    name: "polyfill-node",
    renderChunk(code) {
      return code.replace(/require\(["']fs["']\)/g, "{ readFileSync() {} }");
    },
  };
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
function NodeGlobalsPolyfillPlugin(options) {
  return {
    name: "node-globals-polyfill",
    resolveId(id) {
      if (id === "fs" || id === "crypto") {
        return id;
      }
      return null;
    },
    load(id) {
      if (id === "fs" || id === "crypto") {
        return "export default {}";
      }
      return null;
    },
  };
}

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    Vue(),
    vuetify({
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
