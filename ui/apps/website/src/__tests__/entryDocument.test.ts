import { describe, it, expect } from "vitest";
import indexHtml from "../../index.html?raw";

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
