import { fileURLToPath } from "node:url";
import { defineConfig } from "astro/config";
import { agents } from "../../../src/lib/agents/integration.ts";

const sections = [
  {
    label: "Guides",
    description: "How to do things.",
    items: [
      { label: "Guide", href: "/guide" },
      { label: "FAQ", href: "/about/faq" },
    ],
  },
];

export default defineConfig({
  build: { format: "file" },
  integrations: [agents({ channel: "unreleased", sections })],
  vite: {
    resolve: {
      alias: {
        "@": fileURLToPath(new URL("../../../src", import.meta.url)),
      },
    },
  },
});
