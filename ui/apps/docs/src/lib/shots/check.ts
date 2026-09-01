import { existsSync, readFileSync, readdirSync, rmSync } from "node:fs";
import { extname, join } from "node:path";
import { SHOTS_DIR, shotImageFile } from "./paths";

/** Which of `ids` have no PNG on disk yet. */
export function missingShotImages(publicDir: string, ids: string[]): string[] {
  return ids.filter((id) => !existsSync(join(publicDir, shotImageFile(id))));
}

/**
 * The image directories the docs own outright. Everything in them exists to be
 * shown by a page, so anything no page asks for is a leftover.
 */
const OWNED_DIRS = [SHOTS_DIR, "img/manual"];

const TEXT = new Set([".html", ".css", ".js", ".json", ".xml", ".txt"]);

function* walk(dir: string): Generator<string> {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const path = join(dir, entry.name);

    if (entry.isDirectory()) yield* walk(path);
    else yield path;
  }
}

/**
 * Every image path the built site asks for.
 *
 * Read from the output rather than from the sources, because that is the only
 * answer covering every way a page can reference one: a <Shot> tag, a markdown
 * image, a raw <img>. Astro copies public/ verbatim and rewrites none of these
 * URLs, so what is in the HTML is what a reader will request.
 */
function referencedImages(outDir: string): Set<string> {
  const found = new Set<string>();
  if (!existsSync(outDir)) return found;

  for (const file of walk(outDir)) {
    if (!TEXT.has(extname(file))) continue;

    for (const [path] of readFileSync(file, "utf-8").matchAll(/\/img\/[\w./-]+/g)) {
      found.add(path);
    }
  }

  return found;
}

/** Files in the owned directories that the built site never asks for. */
export function unreferencedImages(publicDir: string, outDir: string): string[] {
  const referenced = referencedImages(outDir);

  // A build that referenced nothing at all is a build that produced no pages,
  // not a site that stopped using every image it has. Proposing the whole
  // directory for deletion on that evidence is the one mistake here that would
  // be expensive.
  if (referenced.size === 0) return [];

  const orphans: string[] = [];

  for (const dir of OWNED_DIRS) {
    const path = join(publicDir, dir);
    if (!existsSync(path)) continue;

    for (const file of readdirSync(path)) {
      if (!referenced.has(`/${dir}/${file}`)) orphans.push(`${dir}/${file}`);
    }
  }

  return orphans.sort();
}

/** Delete files under `publicDir`, tolerating ones already gone. */
export function removeImages(publicDir: string, files: string[]): void {
  for (const file of files) rmSync(join(publicDir, file), { force: true });
}

/**
 * Same split as the registry: the integration and the component load in
 * different module graphs, and Vite inlines `process.env` reads during the SSR
 * transform, so the dev server's state has to travel on `globalThis`. The
 * integration supplies `publicDir` because only it knows which project is being
 * served: resolving it relative to the component would resolve against this app
 * even when another one is building.
 */
const DEV: unique symbol = Symbol.for("@shellhub/docs:shots:dev");

type Host = typeof globalThis & { [DEV]?: { publicDir: string } };

/** Hand the dev server the public directory the render-time check reads. */
export function serveShotsInDev(publicDir: string): void {
  (globalThis as Host)[DEV] = { publicDir };
}

const warned = new Set<string>();

/**
 * The dev server never fires `astro:build:done`, so an author writing a page
 * would get no signal until they ran a build. Warning from the render is the
 * only hook `astro dev` runs.
 */
export function warnMissingShotImageInDev(id: string): void {
  const dev = (globalThis as Host)[DEV];
  if (!dev || warned.has(id)) return;
  if (missingShotImages(dev.publicDir, [id]).length === 0) return;

  warned.add(id);
  console.warn(`[shots] no image captured yet for: ${id}`);
}
