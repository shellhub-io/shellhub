import { existsSync } from "node:fs";
import { join } from "node:path";
import { shotImageFile } from "./paths";

export function missingShotImages(publicDir: string, ids: string[]): string[] {
  return ids.filter((id) => !existsSync(join(publicDir, shotImageFile(id))));
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
