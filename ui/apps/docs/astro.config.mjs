import { defineConfig } from "astro/config";
import mdx from "@astrojs/mdx";
import react from "@astrojs/react";
import sitemap from "@astrojs/sitemap";
import { sidebar } from "./src/data/sidebar.ts";
import { agents } from "./src/lib/agents/integration.ts";
import { SITE, isIndexable } from "./src/lib/seo.ts";
import { shots } from "./src/lib/shots/integration.ts";

export default defineConfig({
  site: SITE,
  trailingSlash: "never",
  build: { format: "file" },
  server: { port: 8083, host: true, allowedHosts: true },
  devToolbar: { enabled: false },
  integrations: [
    mdx(),
    react(),
    sitemap({ filter: (page) => isIndexable(new URL(page).pathname) }),
    shots(),
    agents({ channel: process.env.PUBLIC_DOCS_CHANNEL, sections: sidebar }),
  ],
  compressHTML: true,
});
