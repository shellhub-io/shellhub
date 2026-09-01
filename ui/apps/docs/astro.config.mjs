import { defineConfig } from "astro/config";
import mdx from "@astrojs/mdx";
import react from "@astrojs/react";
import { shots } from "./src/lib/shots/integration.ts";

export default defineConfig({
  server: { port: 8083, host: true, allowedHosts: true },
  devToolbar: { enabled: false },
  integrations: [mdx(), react(), shots()],
  compressHTML: true,
});
