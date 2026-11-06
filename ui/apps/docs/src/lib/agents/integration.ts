import {
  appendFileSync,
  existsSync,
  readFileSync,
  writeFileSync,
} from "node:fs";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import type { AstroIntegration } from "astro";
import type { SidebarSection } from "@/data/sidebar";
import {
  absoluteUrl,
  isIndexable,
  isStableChannel,
  markdownPath,
  normalizePath,
} from "../seo";
import { withFaqSchema } from "./faq";
import { llmsFullTxt, llmsTxt } from "./llms";
import { renderPage, type RenderedPage } from "./markdown";

const NOINDEX_HEADERS = `
/*
  x-robots-tag: noindex
`;

const builtFile = (root: string, path: string): string | undefined => {
  const file = join(root, path === "/" ? "index.html" : `${path}.html`);
  return existsSync(file) ? file : undefined;
};

/**
 * The Astro integration that makes the built site readable by agents. After the build it writes a
 * Markdown copy of every documentation page beside its HTML (`/connect` → `/connect.md`), then
 * `/llms.txt` and `/llms-full.txt` from those copies in the order of `sections`, and publishes the questions of a page marked
 * `schema: faq` as FAQPage structured data. A build of any channel but `stable` also
 * sends `x-robots-tag: noindex` on every response, Markdown and text included.
 */
export function agents({
  channel,
  sections,
}: {
  channel: string | undefined;
  sections: SidebarSection[];
}): AstroIntegration {
  return {
    name: "@shellhub/docs:agents",
    hooks: {
      "astro:build:done": ({ dir, pages, logger }) => {
        const root = fileURLToPath(dir);
        const rendered = new Map<string, RenderedPage>();

        for (const { pathname } of pages) {
          const path = normalizePath(`/${pathname}`);
          const file = builtFile(root, path);
          if (!isIndexable(path) || !file) continue;

          const html = readFileSync(file, "utf-8");
          const withSchema = withFaqSchema(html, path);
          if (withSchema !== html) writeFileSync(file, withSchema);

          const page = renderPage(html, absoluteUrl(path));
          if (!page) continue;

          writeFileSync(join(root, markdownPath(path)), page.markdown);
          rendered.set(path, page);
        }

        writeFileSync(join(root, "llms.txt"), llmsTxt(sections, rendered));
        writeFileSync(
          join(root, "llms-full.txt"),
          llmsFullTxt(sections, rendered),
        );
        logger.info(
          `wrote ${rendered.size} Markdown pages, llms.txt and llms-full.txt`,
        );

        if (!isStableChannel(channel)) {
          appendFileSync(join(root, "_headers"), NOINDEX_HEADERS);
          logger.info(
            "marked every response noindex: this is not a stable build",
          );
        }
      },
    },
  };
}
