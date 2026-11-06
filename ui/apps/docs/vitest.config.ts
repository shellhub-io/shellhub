import { defineConfig } from "vitest/config";
import * as path from "node:path";

export default defineConfig({
  test: {
    environment: "node",
    provide: {
      fixtures: path.resolve(import.meta.dirname, "test/fixtures"),
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "src"),
    },
  },
});
