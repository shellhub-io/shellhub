import { defineConfig } from "astro/config";
import mdx from "@astrojs/mdx";
import react from "@astrojs/react";
import sitemap from "@astrojs/sitemap";
import { shots } from "./src/lib/shots/integration.ts";

export default defineConfig({
  site: "https://docs.shellhub.io",
  server: { port: 8083, host: true, allowedHosts: true },
  devToolbar: { enabled: false },
  integrations: [mdx(), react(), sitemap(), shots()],
  compressHTML: true,
});
