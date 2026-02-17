import { defineConfig } from "astro/config";
import mdx from "@astrojs/mdx";
import tailwind from "@astrojs/tailwind";
import react from "@astrojs/react";

export default defineConfig({
  server: { port: 8084, host: true, allowedHosts: true },
  devToolbar: { enabled: false },
  vite: {
    base: "/v2/blog/",
  },
  integrations: [mdx(), tailwind(), react()],
});
