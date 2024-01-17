import { defineConfig } from "vite";
import Vue from "@vitejs/plugin-vue";
import VuetifyPlugin from "vite-plugin-vuetify";
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
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          dashboard: ["./src/views/Dashboard.vue"],
          devices: [
            "./src/views/Devices.vue",
            "./src/components/Devices/DeviceList.vue",
            "./src/components/Devices/DevicePendingList.vue",
            "./src/components/Devices/DeviceRejectedList.vue",
          ],
          "details-device": ["./src/views/DetailsDevice.vue"],
          sessions: ["./src/views/Sessions.vue"],
          "details-sessions": ["./src/views/DetailsSessions.vue"],
          "firewall-rules": ["./src/views/FirewallRules.vue"],
          "public-keys": ["./src/views/PublicKeys.vue"],
          settings: [
            "./src/views/Settings.vue",
            "./src/components/Setting/SettingProfile.vue",
            "./src/components/Setting/SettingNamespace.vue",
            "./src/components/Setting/SettingPrivateKeys.vue",
            "./src/components/Setting/SettingTags.vue",
            "./src/components/Setting/SettingBilling.vue",
          ],
        },
      },
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
