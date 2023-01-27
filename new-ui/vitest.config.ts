import { defineConfig } from "vitest/config";
import vue from "@vitejs/plugin-vue";
import vuetify from "vite-plugin-vuetify";

import { fileURLToPath, URL } from "node:url";

export default defineConfig({
  // Resolver
  resolve: {
    // https://vitejs.dev/config/#resolve-alias
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  // plugins
  plugins: [
    {
      name: "vitest-plugin-beforeall",
      config: () => ({
        test: {
          setupFiles: [
            fileURLToPath(new URL("./vitest/beforeAll.ts", import.meta.url)),
          ],
        },
      }),
    } as any,
    // Vue3
    vue(),
    // Vuetify Loader
    // https://github.com/vuetifyjs/vuetify-loader
    vuetify({
      autoImport: true,
      styles: { configFile: "./src/styles/variables.scss" },
    }),
  ],
  test: {
    // https://vitest.dev/guide/#configuring-vitest
    globals: true,
    globalSetup: [fileURLToPath(new URL("./vitest/setup.ts", import.meta.url))],
    setupFiles: ["./vitest/vitest-canvas-mock.ts"],
    environment: "jsdom",
    deps: {
      inline: ["vuetify"],
    },
    exclude: ["**/node_modules/**"],
    update: true,
    threads: false,
  },
});
