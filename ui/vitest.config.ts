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
      "@admin": fileURLToPath(new URL("./admin/src", import.meta.url)),
    },
  },
  css: {
    preprocessorOptions: {
      sass: {
        api: "modern-compiler",
        silenceDeprecations: ["legacy-js-api"],
      },
    },
  },
  // plugins
  plugins: [
    vue(),
    vuetify({
      autoImport: true,
      styles: { configFile: "./src/styles/variables.scss" },
    }),
    {
      name: "vitest-plugin-beforeall",
      config: () => ({
        test: {
          setupFiles: [
            fileURLToPath(new URL("./vitest/beforeAll.ts", import.meta.url)),
          ],
        },
      }),
    },
  ],
  test: {
    // https://vitest.dev/guide/#configuring-vitest
    globals: true,
    globalSetup: [fileURLToPath(new URL("./vitest/setup.ts", import.meta.url))],
    setupFiles: ["./vitest/vitest-canvas-mock.ts"],
    environment: "jsdom",
    server: {
      deps: {
        inline: ["vuetify"],
      },
    },
    exclude: ["**/node_modules/**"],
    update: false,
    // onConsoleLog(log: string): boolean | void {
    //   return !(log.includes("Axios"));
    // },
  },
});
