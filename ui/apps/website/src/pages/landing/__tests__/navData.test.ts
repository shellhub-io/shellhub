import { describe, it, expect } from "vitest";
import { navHrefs } from "@/pages/landing/navData";
import { routes } from "@/routes";

// Only internal paths (starting with "/") need to resolve to a known route.
// "#" placeholders are intentional stubs (changelog, blog, discord, etc.) that
// have no corresponding page yet — skipping them prevents false positives.
// "http*" URLs point to external sites (docs, GitHub) and are outside our
// router's jurisdiction — the browser handles them, not React Router.
const routePaths = new Set(routes.map((r) => r.path));

const internalHrefs = navHrefs.filter(
  (href) => href.startsWith("/") && href !== "#",
);

describe("nav dead-link integrity", () => {
  it("every internal nav href resolves to a known route", () => {
    const deadLinks = internalHrefs.filter((href) => !routePaths.has(href));
    expect(
      deadLinks,
      `Dead links found in nav: ${deadLinks.join(", ")}`,
    ).toHaveLength(0);
  });

  // /pricing lives exclusively in the simpleLinks array inside navData.tsx.
  // The dead-link test above only proves that *present* hrefs resolve — it
  // would stay green even if /pricing were accidentally dropped from
  // simpleLinks (and thus from navHrefs). This guard fails loudly in that
  // case, ensuring the pricing page is never silently removed from the nav.
  it("/pricing is present in navHrefs", () => {
    expect(
      navHrefs,
      "/pricing must remain in navHrefs (sourced from simpleLinks)",
    ).toContain("/pricing");
  });
});
