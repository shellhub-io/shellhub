import { fileURLToPath } from "node:url";
import mdx from "@astrojs/mdx";
import { defineConfig } from "astro/config";
import { shots } from "../../../src/lib/shots/integration.ts";

export default defineConfig({
  integrations: [mdx(), shots()],
  vite: {
    resolve: {
      alias: {
        "@": fileURLToPath(new URL("../../../src", import.meta.url)),
      },
    },
  },
});
