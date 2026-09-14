import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { resolve, relative, sep } from "node:path";
import { fileURLToPath } from "node:url";
import { JSDOM } from "jsdom";

const localOrigin = "https://docs.shellhub.io";
const resourceSelector = [
  "a[href]",
  "area[href]",
  "audio[src]",
  "iframe[src]",
  "img[src]",
  "link[href]",
  "script[src]",
  "source[src]",
  "video[src]",
].join(",");

function filesUnder(directory) {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = resolve(directory, entry.name);
    return entry.isDirectory() ? filesUnder(path) : [path];
  });
}

function routeFor(root, file) {
  const path = relative(root, file).split(sep).join("/");
  if (path === "index.html") return "/";
  if (path.endsWith("/index.html")) return `/${path.slice(0, -10)}`;
  return `/${path}`;
}

function targetFor(root, pathname) {
  const decoded = decodeURIComponent(pathname).replace(/^\/+/, "");
  const direct = resolve(root, decoded);
  const candidates = pathname.endsWith("/")
    ? [resolve(direct, "index.html")]
    : [direct, `${direct}.html`, resolve(direct, "index.html")];
  return candidates.find((candidate) => existsSync(candidate) && statSync(candidate).isFile());
}

export function checkLinks(rootDirectory) {
  const root = resolve(rootDirectory);
  const htmlFiles = filesUnder(root).filter((file) => file.endsWith(".html"));
  const documents = new Map();
  const errors = [];

  const documentFor = (file) => {
    if (!documents.has(file)) {
      documents.set(file, new JSDOM(readFileSync(file, "utf8")).window.document);
    }
    return documents.get(file);
  };

  for (const file of htmlFiles) {
    const document = documentFor(file);
    const route = routeFor(root, file);

    for (const element of document.querySelectorAll(resourceSelector)) {
      const attribute = element.hasAttribute("href") ? "href" : "src";
      const value = element.getAttribute(attribute)?.trim();
      if (!value) continue;

      const url = new URL(value, `${localOrigin}${route}`);
      if (url.origin !== localOrigin) continue;

      const target = targetFor(root, url.pathname);
      if (!target) {
        errors.push(`${route}: ${value} does not exist`);
        continue;
      }

      if (url.hash && target.endsWith(".html")) {
        const id = decodeURIComponent(url.hash.slice(1));
        const ids = new Set([...documentFor(target).querySelectorAll("[id]")].map((node) => node.id));
        if (id && !ids.has(id)) errors.push(`${route}: ${value} has no matching fragment`);
      }
    }
  }

  return [...new Set(errors)].sort();
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const errors = checkLinks(process.argv[2] ?? "dist");
  if (errors.length > 0) {
    for (const error of errors) console.error(error);
    process.exitCode = 1;
  }
}
