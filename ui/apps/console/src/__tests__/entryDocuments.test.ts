import { describe, it, expect } from "vitest";
import indexHtml from "../../index.html?raw";
import migrateHtml from "../../migrate.html?raw";

/**
 * Chrome's auto-translation moves text nodes into <font> wrappers, which detaches the text
 * nodes React tracks and makes the next insertBefore/removeChild throw NotFoundError.
 * Opting the whole document out is the only defence that covers every route at once.
 */
const ENTRY_POINTS: [string, string][] = [
  ["index.html", indexHtml],
  ["migrate.html", migrateHtml],
];

describe.each(ENTRY_POINTS)("%s", (_name, html) => {
  const doc = new DOMParser().parseFromString(html, "text/html");

  it("opts the document out of browser auto-translation", () => {
    expect(doc.documentElement.getAttribute("translate")).toBe("no");
  });

  it("opts out of Google Translate's page-level toolbar", () => {
    const meta = doc.querySelector('meta[name="google"]');

    expect(meta?.getAttribute("content")).toBe("notranslate");
  });
});
