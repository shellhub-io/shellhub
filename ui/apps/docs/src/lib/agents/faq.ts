import type { ElementContent } from "hast";
import { fromHtml } from "hast-util-from-html";
import { select } from "hast-util-select";
import { toText } from "hast-util-to-text";
import { faqGraph, jsonLd, type Faq } from "../seo";

const collapse = (text: string): string => text.replace(/\s+/g, " ").trim();

const isHeading = (node: ElementContent, levels: string[]): boolean =>
  node.type === "element" && levels.includes(node.tagName);

/**
 * Reads a built page marked `data-schema="faq"` as questions and answers: each `h2` of its article
 * is a question, and everything up to the next `h1` or `h2` is the answer. Returns an empty list
 * for a page without the mark. A question with no answer under it is skipped.
 */
export const extractFaq = (html: string): Faq[] => {
  const article = select('article.prose[data-schema="faq"]', fromHtml(html));
  if (!article) return [];

  const nodes = article.children;

  return nodes.flatMap((node, index) => {
    if (!isHeading(node, ["h2"])) return [];

    const end = nodes.findIndex(
      (next, at) => at > index && isHeading(next, ["h1", "h2"]),
    );
    const body = nodes.slice(index + 1, end === -1 ? undefined : end);
    const answer = collapse(body.map((part) => toText(part)).join(" "));

    return answer ? [{ question: collapse(toText(node)), answer }] : [];
  });
};

/**
 * Returns the page with its questions published as FAQPage structured data in its `<head>`, or
 * the page unchanged when `extractFaq` finds none.
 */
export const withFaqSchema = (html: string, pathname: string): string => {
  const faq = extractFaq(html);
  if (faq.length === 0) return html;

  const script = `<script type="application/ld+json">${jsonLd(faqGraph(pathname, faq))}</script>`;
  return html.replace("</head>", () => `${script}</head>`);
};
