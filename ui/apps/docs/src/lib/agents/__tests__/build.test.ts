import { type ChildProcess, execFileSync, spawn } from "node:child_process";
import { existsSync, readFileSync, rmSync } from "node:fs";
import { join } from "node:path";
import { fromHtml } from "hast-util-from-html";
import { select } from "hast-util-select";
import type { Link } from "mdast";
import { afterAll, beforeAll, describe, expect, inject, it } from "vitest";
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

const spawnDev = (root: string): ChildProcess => {
  const { VITEST: _vitest, ...env } = process.env;

  return spawn("npx", ["astro", "dev", "--root", ".", "--host", "127.0.0.1", "--port", "0"], {
    cwd: root,
    detached: true,
    env: { ...env, CI: "", NO_COLOR: "1" },
  });
};

const originOf = (server: ChildProcess): Promise<string> =>
  new Promise((resolve, reject) => {
    let output = "";
    const read = (chunk: Buffer) => {
      output += chunk.toString();
      const origin = /http:\/\/127\.0\.0\.1:\d+/.exec(output)?.[0];
      if (origin) resolve(origin);
    };

    server.stdout?.on("data", read);
    server.stderr?.on("data", read);
    server.on("exit", (code) => reject(new Error(`astro dev exited with ${code}:\n${output}`)));
  });

const stop = async (server: ChildProcess | undefined) => {
  if (!server?.pid || server.exitCode !== null) return;

  const exited = new Promise((resolve) => server.once("exit", resolve));
  process.kill(-server.pid, "SIGTERM");
  await exited;
};

describe("agents integration, over the astro dev server", () => {
  let server: ChildProcess | undefined;
  let origin = "";
  const get = (path: string) => fetch(`${origin}${path}`);

  beforeAll(async () => {
    server = spawnDev(fixture);
    origin = await originOf(server);
  }, 120_000);

  afterAll(async () => {
    await stop(server);
  });

  it("serves a page's Markdown copy before any build", async () => {
    const response = await get("/guide.md");
    const links = nodesOf<Link>(parse(await response.text()), "link").map((node) => node.url);

    expect(response.headers.get("content-type")).toBe("text/markdown; charset=utf-8");
    expect(links).toEqual(["https://docs.shellhub.io/about/faq"]);
  });

  it("serves no copy of the 404 page", async () => {
    const response = await get("/404.md");

    expect(response.status).toBe(404);
    expect(response.headers.get("content-type")).not.toMatch(/^text\/markdown/);
  });
});
