import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { removeImages, unreferencedImages } from "@/lib/shots/check";

let root = "";
const publicDir = () => join(root, "public");
const outDir = () => join(root, "dist");

function image(dir: string, name: string): void {
  mkdirSync(join(publicDir(), dir), { recursive: true });
  writeFileSync(join(publicDir(), dir, name), "");
}

function page(html: string): void {
  mkdirSync(join(outDir(), "guides"), { recursive: true });
  writeFileSync(join(outDir(), "guides", "index.html"), html);
}

beforeEach(() => {
  root = mkdtempSync(join(tmpdir(), "shots-check-"));
});

afterEach(() => {
  rmSync(root, { recursive: true, force: true });
});

describe("unreferencedImages", () => {
  it("keeps a captured shot the built page displays", () => {
    image("img/shots", "device-list.png");
    page('<img src="/img/shots/device-list.png">');

    expect(unreferencedImages(publicDir(), outDir())).toEqual([]);
  });

  it("reports a shot no page displays any more", () => {
    image("img/shots", "device-list.png");
    image("img/shots", "firewall-rules.png");
    page('<img src="/img/shots/device-list.png">');

    expect(unreferencedImages(publicDir(), outDir())).toEqual([
      "img/shots/firewall-rules.png",
    ]);
  });

  // Manual images are referenced by plain markdown, which no manifest knows
  // about. Reading the built html is what lets one rule cover both directories.
  it("reports a manual image no page displays any more", () => {
    image("img/manual", "mfa-setup.png");
    image("img/manual", "public-keys.png");
    page('<img src="/img/manual/mfa-setup.png" alt="setup">');

    expect(unreferencedImages(publicDir(), outDir())).toEqual([
      "img/manual/public-keys.png",
    ]);
  });

  it("counts a reference from anywhere in the output", () => {
    image("img/manual", "session-play.gif");
    mkdirSync(outDir(), { recursive: true });
    writeFileSync(join(outDir(), "styles.css"), "a{background:url(/img/manual/session-play.gif)}");

    expect(unreferencedImages(publicDir(), outDir())).toEqual([]);
  });

  // Every image lives in one of the owned directories, so a build that wrote no
  // html would otherwise propose deleting all of them at once.
  it("proposes nothing when there is no build to read", () => {
    image("img/shots", "device-list.png");

    expect(unreferencedImages(publicDir(), join(root, "nowhere"))).toEqual([]);
  });

  it("ignores directories the docs do not own", () => {
    image("img/logos", "shellhub.svg");
    page("<p>no images here</p>");

    expect(unreferencedImages(publicDir(), outDir())).toEqual([]);
  });
});

describe("removeImages", () => {
  it("deletes exactly what it was given", () => {
    image("img/shots", "device-list.png");
    image("img/manual", "public-keys.png");
    page('<img src="/img/shots/device-list.png">');

    removeImages(publicDir(), unreferencedImages(publicDir(), outDir()));

    expect(unreferencedImages(publicDir(), outDir())).toEqual([]);
    expect(existsSync(join(publicDir(), "img/shots/device-list.png"))).toBe(true);
    expect(existsSync(join(publicDir(), "img/manual/public-keys.png"))).toBe(false);
  });
});
