import { execFileSync } from "node:child_process";
import { readFileSync, rmSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { JSDOM } from "jsdom";
import { beforeAll, describe, expect, it } from "vitest";
import type { ShotManifest } from "@/lib/shots/manifest";

const fixture = fileURLToPath(
  new URL("../../../../test/fixtures/shots-site", import.meta.url),
);

let manifest: ShotManifest;

/**
 * The component runs in Vite's SSR module graph and the integration in Node's.
 * Those are separate module caches, so a registry that works under vitest can
 * still hand the integration an empty array during a real build. Only an actual
 * `astro build` exercises both graphs, which is why this one test is worth its
 * cost.
 */
describe("shots integration, over a real astro build", () => {
  beforeAll(() => {
    rmSync(`${fixture}/dist`, { recursive: true, force: true });
    rmSync(`${fixture}/.astro`, { recursive: true, force: true });

    execFileSync("npx", ["astro", "build", "--root", "."], {
      cwd: fixture,
      encoding: "utf-8",
      env: { ...process.env, CI: "" },
    });

    manifest = JSON.parse(
      readFileSync(`${fixture}/.astro/shots.json`, "utf-8"),
    ) as ShotManifest;
  }, 120_000);

  it("emits every declared shot, merging the one used on both pages", () => {
    expect(manifest.shots).toEqual([
      {
        id: "dashboard",
        route: "/",
        viewport: { width: 800, height: 600 },
        edition: "ce",
        usedBy: ["/one/"],
      },
      {
        id: "device-list",
        route: "/devices",
        viewport: { width: 1440, height: 900 },
        edition: "ce",
        of: { role: "table" },
        usedBy: ["/one/", "/two/"],
      },
      {
        id: "session-list",
        route: "/sessions",
        viewport: { width: 1440, height: 900 },
        edition: "enterprise",
        usedBy: ["/two/"],
      },
    ]);
  });

  it("renders the image that displays each shot", () => {
    const html = readFileSync(`${fixture}/dist/one/index.html`, "utf-8");
    const { document } = new JSDOM(html).window;

    const images = [...document.querySelectorAll("img")].map((image) => ({
      src: image.getAttribute("src"),
      alt: image.getAttribute("alt"),
    }));

    expect(images).toContainEqual({
      src: "/img/shots/device-list.png",
      alt: "Devices",
    });
  });
});
