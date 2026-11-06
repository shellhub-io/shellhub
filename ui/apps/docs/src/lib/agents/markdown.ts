import type { Element, ElementContent } from "hast";
import { fromHtml } from "hast-util-from-html";
import { matches, select, selectAll } from "hast-util-select";
import { toMdast } from "hast-util-to-mdast";
import { toText } from "hast-util-to-text";
import { gfmToMarkdown } from "mdast-util-gfm";
import { toMarkdown } from "mdast-util-to-markdown";

/** A built page read back as what an agent needs: what it is called, what it says, and where. */
export interface RenderedPage {
  title: string;
  description: string;
  url: string;
  markdown: string;
}

const NOISE = [
  "script",
  "style",
  "template",
  "noscript",
  "button",
  "input",
  "select",
  "dialog",
  "svg",
  ".replay-bar",
  "[hidden]",
  "[aria-hidden=true]",
].join(", ");

const text = (node: Element | undefined, property: string): string => {
  const value = node?.properties[property];
  return typeof value === "string" ? value.trim() : "";
};

const element = (tagName: string, children: ElementContent[]): Element => ({
  type: "element",
  tagName,
  properties: {},
  children,
});

const replace = (
  tree: Element,
  test: (node: Element) => boolean,
  by: (node: Element) => ElementContent[],
): void => {
  tree.children = tree.children.flatMap<ElementContent>((child) => {
    if (child.type !== "element") return [child];
    if (test(child)) return by(child);

    replace(child, test, by);
    return [child];
  });
};

const absolutize = (article: Element, base: string): void => {
  for (const [selector, property] of [
    ["a[href]", "href"],
    ["img[src]", "src"],
    ["source[src]", "src"],
  ] as const) {
    for (const node of selectAll(selector, article)) {
      const value = text(node, property);
      if (value) node.properties[property] = new URL(value, base).href;
    }
  }
};

const labelCodeBlocks = (article: Element): void => {
  for (const pre of selectAll("pre[data-language]", article)) {
    const language = text(pre, "dataLanguage");
    const code = select("code", pre);
    if (code && language && language !== "plaintext")
      code.properties.className = [`language-${language}`];
  }
};

const describeFigures = (article: Element): void => {
  replace(
    article,
    (node) =>
      node.properties.role === "img" && Boolean(text(node, "ariaLabel")),
    (node) => [
      element("p", [
        element("em", [{ type: "text", value: text(node, "ariaLabel") }]),
      ]),
    ],
  );
};

const labelClips = (article: Element): void => {
  for (const video of selectAll("video[aria-label]", article)) {
    video.children.push({ type: "text", value: text(video, "ariaLabel") });
  }
};

const quoteCallouts = (article: Element): void => {
  for (const callout of selectAll("aside.callout", article)) {
    callout.tagName = "blockquote";
    const head = select(".callout-head", callout);
    if (head)
      head.children = [
        element("strong", [{ type: "text", value: toText(head).trim() }]),
      ];
  }
};

/**
 * Reads a page the site built and returns its article as Markdown, with every link and image made
 * absolute against `url`. Returns undefined for a page with no `article.prose`, which is every page
 * that is not set in the documentation layout: the home and the API reference.
 */
export const renderPage = (
  html: string,
  url: string,
): RenderedPage | undefined => {
  const tree = fromHtml(html);
  const article = select("article.prose", tree);
  if (!article) return undefined;

  const heading = select("h1", article);
  const title = (
    heading ? toText(heading) : toText(select("title", tree) ?? tree)
  ).trim();
  const description = text(select("meta[name=description]", tree), "content");

  describeFigures(article);
  replace(
    article,
    (node) => node.properties.dataShotZoom !== undefined,
    (node) => node.children,
  );
  replace(
    article,
    (node) => matches(NOISE, node),
    () => [],
  );
  labelClips(article);
  quoteCallouts(article);
  labelCodeBlocks(article);
  absolutize(article, url);

  const markdown = toMarkdown(
    toMdast({ type: "root", children: article.children }),
    {
      bullet: "-",
      emphasis: "_",
      fences: true,
      rule: "-",
      extensions: [gfmToMarkdown()],
    },
  ).trim();

  return { title, description, url, markdown: `${markdown}\n` };
};
