import { describe, it, expect } from "vitest";
import indexHtml from "../../index.html?raw";

// See the console app's entryDocuments test: auto-translation detaches the text nodes React
// tracks, so every React document opts out at the root.
describe("index.html", () => {
  const doc = new DOMParser().parseFromString(indexHtml, "text/html");

  it("opts the document out of browser auto-translation", () => {
    expect(doc.documentElement.getAttribute("translate")).toBe("no");
  });

  it("opts out of Google Translate's page-level toolbar", () => {
    const meta = doc.querySelector('meta[name="google"]');

    expect(meta?.getAttribute("content")).toBe("notranslate");
  });
});
