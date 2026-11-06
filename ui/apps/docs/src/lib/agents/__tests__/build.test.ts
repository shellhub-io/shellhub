import { execFileSync } from "node:child_process";
import { existsSync, readFileSync, rmSync } from "node:fs";
import { join } from "node:path";
import { fromHtml } from "hast-util-from-html";
import { select } from "hast-util-select";
import type { Link } from "mdast";
import { beforeAll, describe, expect, inject, it } from "vitest";
import { nodesOf, parse } from "./mdast";

declare module "vitest" {
  export interface ProvidedContext {
    fixtures: string;
  }
}

const fixture = join(inject("fixtures"), "agents-site");
const dist = join(fixture, "dist");

const read = (file: string): string => readFileSync(join(dist, file), "utf-8");

describe("agents integration, over a real astro build of the unreleased channel", () => {
  beforeAll(() => {
    rmSync(dist, { recursive: true, force: true });
    rmSync(join(fixture, ".astro"), { recursive: true, force: true });

    execFileSync("npx", ["astro", "build", "--root", "."], {
      cwd: fixture,
      encoding: "utf-8",
      env: { ...process.env, CI: "" },
    });
  }, 120_000);

  it("writes a Markdown copy beside each page, and none for the 404", () => {
    expect(existsSync(join(dist, "guide.md"))).toBe(true);
    expect(existsSync(join(dist, "about/faq.md"))).toBe(true);
    expect(existsSync(join(dist, "404.md"))).toBe(false);
  });

  it("lists the pages in llms.txt in the order of the sections it was given", () => {
    const urls = nodesOf<Link>(parse(read("llms.txt")), "link").map((node) => node.url);

    expect(urls.slice(0, 2)).toEqual([
      "https://docs.shellhub.io/guide.md",
      "https://docs.shellhub.io/about/faq.md",
    ]);
  });

  it("adds FAQPage data to the marked page only", () => {
    const schema = (file: string): unknown[] => {
      const script = select('head script[type="application/ld+json"]', fromHtml(read(file)));
      const text = script?.children[0];

      return text?.type === "text" ? [JSON.parse(text.value) as unknown] : [];
    };

    expect(schema("about/faq.html")).toMatchObject([
      { "@type": "FAQPage", mainEntity: [{ name: "Is it a VPN?", acceptedAnswer: { text: "No." } }] },
    ]);
    expect(schema("guide.html")).toEqual([]);
  });

  it("adds a noindex rule to the site's own headers, keeping the rules already there", () => {
    const rules = read("_headers")
      .split("\n")
      .reduce<{ path: string; headers: string[] }[]>((all, line) => {
        if (line.trim() === "") return all;
        if (!/^\s/.test(line)) return [...all, { path: line.trim(), headers: [] }];

        all.at(-1)?.headers.push(line.trim());
        return all;
      }, []);

    expect(rules).toEqual([
      { path: "/*", headers: ["x-content-type-options: nosniff"] },
      { path: "/*", headers: ["x-robots-tag: noindex"] },
    ]);
  });
});
