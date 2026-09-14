import { readdirSync } from "node:fs";
import { join } from "node:path";

const PAGES_DIR = join(import.meta.dirname, "../pages");

/**
 * Every address the site serves, derived from the files under `src/pages` the way Astro routes
 * them: the extension goes, an `index` becomes its directory, and the root becomes `/`.
 */
export const routes = (): Set<string> =>
  new Set(
    readdirSync(PAGES_DIR, { recursive: true, encoding: "utf-8" })
      .filter((file) => /\.(mdx|astro)$/.test(file))
      .map(
        (file) =>
          ("/" + file.replace(/\.(mdx|astro)$/, "")).replace(/\/index$/, "") ||
          "/",
      ),
  );
