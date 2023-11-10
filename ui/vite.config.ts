import { defineConfig } from "vitest/config";
import vue from "@vitejs/plugin-vue";
// https://github.com/vuetifyjs/vuetify-loader/tree/next/packages/vite-plugin
import vuetify, { transformAssetUrls } from "vite-plugin-vuetify";
import NodeGlobalsPolyfillPlugin from "@esbuild-plugins/node-globals-polyfill";
import polyfillNode from "rollup-plugin-polyfill-node";
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
    vue({
      template: { transformAssetUrls },
      script: {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        // Experimental defineModel from Vue 3.3
        defineModel: true,
      },
    }),
    vuetify({
      autoImport: true,
    }),
    Markdown({
      markdownItOptions: {
        html: true,
        typographer: true,
      },
    }),
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
        polyfillNode,
        NodeGlobalsPolyfillPlugin({
          process: true,
          buffer: true,
        }),
      ],
    },
  },
});
