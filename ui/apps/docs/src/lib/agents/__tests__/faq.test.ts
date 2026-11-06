import { fromHtml } from "hast-util-from-html";
import { select } from "hast-util-select";
import { describe, expect, it } from "vitest";
import { extractFaq, withFaqSchema } from "@/lib/agents/faq";

const page = (article: string, schema = ' data-schema="faq"'): string =>
  `<html><head><title>FAQ</title></head><body><article class="prose"${schema}>${article}</article></body></html>`;

describe("extractFaq", () => {
  it("reads each second-level heading as a question and what follows as its answer", () => {
    const html = page(
      "<h1>FAQ</h1><h2>Does it need a port?</h2><p>No. The agent <code>dials</code> out.</p>" +
        "<ul><li>Outbound only</li><li>Port 443</li></ul><h2>Is it a VPN?</h2><p>No.</p>",
    );

    expect(extractFaq(html)).toEqual([
      {
        question: "Does it need a port?",
        answer: "No. The agent dials out. Outbound only Port 443",
      },
      { question: "Is it a VPN?", answer: "No." },
    ]);
  });

  it("skips a question with no answer", () => {
    expect(
      extractFaq(page("<h2>Empty?</h2><h2>Full?</h2><p>Yes.</p>")),
    ).toEqual([{ question: "Full?", answer: "Yes." }]);
  });

  it("reads nothing from a page that is not marked", () => {
    expect(extractFaq(page("<h2>Q?</h2><p>A.</p>", ""))).toEqual([]);
  });
});

describe("withFaqSchema", () => {
  it("adds FAQPage structured data to the head", () => {
    const html = withFaqSchema(page("<h2>Q?</h2><p>A.</p>"), "/about/faq.html");
    const script = select(
      'head script[type="application/ld+json"]',
      fromHtml(html),
    );
    const text = script?.children[0];

    expect(JSON.parse(text?.type === "text" ? text.value : "")).toMatchObject({
      "@type": "FAQPage",
      url: "https://docs.shellhub.io/about/faq",
      mainEntity: [{ name: "Q?", acceptedAnswer: { text: "A." } }],
    });
  });

  it("publishes an answer containing $ patterns verbatim", () => {
    const answer = "Run kill $$ or printf $'a\\tb' and read $&.";
    const html = withFaqSchema(
      page(`<h2>Q?</h2><p>${answer}</p><p>After.</p>`),
      "/about/faq",
    );
    const tree = fromHtml(html);
    const script = select('head script[type="application/ld+json"]', tree);
    const text = script?.children[0];

    expect(JSON.parse(text?.type === "text" ? text.value : "")).toMatchObject({
      mainEntity: [{ acceptedAnswer: { text: `${answer} After.` } }],
    });
  });

  it("leaves a page without questions unchanged", () => {
    const html = page("<p>Nothing to ask.</p>");

    expect(withFaqSchema(html, "/about")).toBe(html);
  });
});
