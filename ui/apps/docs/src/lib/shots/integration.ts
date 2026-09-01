import { mkdirSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import type { AstroIntegration } from "astro";
import {
  missingShotImages,
  removeImages,
  serveShotsInDev,
  unreferencedImages,
} from "./check";
import { buildManifest } from "./manifest";
import { collectShots, resetShots } from "./registry";

const MANIFEST_PATH = ".astro/shots.json";

/**
 * The Astro integration that writes the manifest and guards the image
 * directory: it collects what the pages declared, and reports an id with no
 * image and an image no page asks for.
 */
export function shots(): AstroIntegration {
  let root = "";
  let publicDir = "";
  let command = "";

  return {
    name: "@shellhub/docs:shots",
    hooks: {
      "astro:config:setup": (options) => {
        command = options.command;
      },

      "astro:config:done": ({ config }) => {
        root = fileURLToPath(config.root);
        publicDir = fileURLToPath(config.publicDir);

        if (command === "dev") serveShotsInDev(publicDir);
      },

      "astro:build:start": () => {
        // A watched rebuild reuses the process, so last build's entries survive.
        resetShots();
      },

      "astro:build:done": ({ dir, logger }) => {
        const manifest = buildManifest(collectShots());
        const out = join(root, MANIFEST_PATH);

        mkdirSync(dirname(out), { recursive: true });
        writeFileSync(out, `${JSON.stringify(manifest, null, 2)}\n`);
        logger.info(`wrote ${manifest.shots.length} shots to ${MANIFEST_PATH}`);

        const orphans = unreferencedImages(publicDir, fileURLToPath(dir));

        if (orphans.length > 0) {
          // CI reports what was left behind rather than deleting it: a build
          // that quietly rewrites the checkout hides the fact that the removal
          // was never committed.
          const left = `no page references: ${orphans.join(", ")}`;
          if (process.env.CI) throw new Error(left);

          removeImages(publicDir, orphans);
          logger.info(`removed ${orphans.length} unused: ${orphans.join(", ")}`);
        }

        const missing = missingShotImages(
          publicDir,
          manifest.shots.map((shot) => shot.id),
        );

        if (missing.length === 0) return;

        const message = `no image captured yet for: ${missing.join(", ")}`;

        // A missing image is normal while writing a page and unacceptable on the
        // published site, so it is a warning locally and a failure in CI.
        if (process.env.CI) throw new Error(message);
        logger.warn(message);
      },
    },
  };
}
