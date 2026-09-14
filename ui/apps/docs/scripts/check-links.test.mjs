import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";
import { checkLinks, externalLinks } from "./check-links.mjs";

const roots = [];

afterEach(() => {
  for (const root of roots.splice(0)) rmSync(root, { recursive: true });
});

describe("externalLinks", () => {
  it("returns unique HTTP links without local, mail, or data URLs", () => {
    const root = site({
      "index.html": [
        '<a href="https://example.com/guide">External</a>',
        '<a href="https://example.com/guide">Duplicate</a>',
        '<a href="/guide">Local</a>',
        '<a href="mailto:help@example.com">Mail</a>',
        '<img src="data:image/png;base64,eA==">',
      ].join(""),
    });

    expect(externalLinks(root)).toEqual(["https://example.com/guide"]);
  });
});

function site(files) {
  const root = mkdtempSync(join(tmpdir(), "docs-links-"));
  roots.push(root);
  for (const [path, contents] of Object.entries(files)) {
    const file = join(root, path);
    mkdirSync(join(file, ".."), { recursive: true });
    writeFileSync(file, contents);
  }
  return root;
}

describe("checkLinks", () => {
  it("accepts local routes, fragments, assets, and external URLs", () => {
    const root = site({
      "index.html": '<a href="/guide/#install">Guide</a><a href="https://example.com">External</a>',
      "guide/index.html": '<h1 id="install">Install</h1><img src="/image.png">',
      "image.png": "image",
    });

    expect(checkLinks(root)).toEqual([]);
  });

  it("reports missing routes, fragments, and assets", () => {
    const root = site({
      "index.html": '<a href="/missing">Missing</a><a href="/guide/#missing">Fragment</a>',
      "guide/index.html": '<h1 id="install">Install</h1><img src="/missing.png">',
    });

    expect(checkLinks(root)).toEqual([
      "/: /guide/#missing has no matching fragment",
      "/: /missing does not exist",
      "/guide/: /missing.png does not exist",
    ]);
  });
});
