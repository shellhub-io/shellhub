import { readdirSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";
import type { SidebarItem } from "@/data/sidebar";
import { PAGES_NOT_IN_NAV, flattenItems, sidebar } from "@/data/sidebar";

describe("flattenItems", () => {
  describe("group 1: pure-function tests over hand-crafted inputs", () => {
    it("returns a leaf item with href as-is", () => {
      const items: SidebarItem[] = [{ label: "Leaf", href: "/leaf" }];

      expect(flattenItems(items)).toEqual([
        { label: "Leaf", href: "/leaf", featured: undefined },
      ]);
    });

    it("recurses into a branch that has only items (no href)", () => {
      const items: SidebarItem[] = [
        {
          label: "Branch",
          items: [
            { label: "Child A", href: "/child-a" },
            { label: "Child B", href: "/child-b" },
          ],
        },
      ];

      expect(flattenItems(items)).toEqual([
        { label: "Child A", href: "/child-a", featured: undefined },
        { label: "Child B", href: "/child-b", featured: undefined },
      ]);
    });

    it("flattens a nested branch (branch inside a branch)", () => {
      const items: SidebarItem[] = [
        {
          label: "Outer",
          items: [
            {
              label: "Inner",
              items: [{ label: "Deep Leaf", href: "/deep" }],
            },
          ],
        },
      ];

      expect(flattenItems(items)).toEqual([
        { label: "Deep Leaf", href: "/deep", featured: undefined },
      ]);
    });

    it("returns [] for an empty items array", () => {
      expect(flattenItems([])).toEqual([]);
    });

    it("preserves the featured flag on leaf items", () => {
      const items: SidebarItem[] = [
        { label: "Featured", href: "/featured", featured: true },
        { label: "Normal", href: "/normal" },
      ];

      expect(flattenItems(items)).toEqual([
        { label: "Featured", href: "/featured", featured: true },
        { label: "Normal", href: "/normal", featured: undefined },
      ]);
    });

    // Defensive documentation of the ternary short-circuit: if an item has BOTH
    // href and items, href takes precedence and children are dropped. In practice
    // the real sidebar tree never contains such an item (asserted in group 2),
    // so this is a hypothetical dead path — documented here so the short-circuit
    // behavior is explicit and observable.
    it("short-circuits on href when an item has both href and items (href wins, children dropped)", () => {
      const items: SidebarItem[] = [
        {
          label: "Ambiguous",
          href: "/ambiguous",
          items: [{ label: "Orphaned Child", href: "/orphaned" }],
        },
      ];

      expect(flattenItems(items)).toEqual([
        { label: "Ambiguous", href: "/ambiguous", featured: undefined },
      ]);
    });
  });

  describe("group 2: structural invariants over the real sidebar constant", () => {
    // Helper: recursively walk all items in a section (not just leaves).
    function walkItems(items: SidebarItem[]): SidebarItem[] {
      return items.flatMap((item) =>
        item.items ? [item, ...walkItems(item.items)] : [item],
      );
    }

    it("every section has a non-empty label, description, icon, and at least one item", () => {
      for (const section of sidebar) {
        expect(section.label, "section label must be non-empty").toBeTruthy();
        expect(
          section.description,
          `section "${section.label}" description must be non-empty`,
        ).toBeTruthy();
        expect(
          section.icon,
          `section "${section.label}" icon must be non-empty`,
        ).toBeTruthy();
        expect(
          section.items.length,
          `section "${section.label}" items must be non-empty`,
        ).toBeGreaterThan(0);
      }
    });

    it("every leaf item produced by flattenItems has a non-empty label", () => {
      for (const section of sidebar) {
        const leaves = flattenItems(section.items);
        for (const leaf of leaves) {
          expect(
            leaf.label,
            `leaf in section "${section.label}" must have non-empty label`,
          ).toBeTruthy();
        }
      }
    });

    it("every SidebarItem is either a link (has href) OR a non-empty branch (has items with length > 0) — no dead nodes", () => {
      for (const section of sidebar) {
        for (const item of walkItems(section.items)) {
          const isLink = typeof item.href === "string" && item.href.length > 0;
          const isBranch = Array.isArray(item.items) && item.items.length > 0;
          expect(
            isLink || isBranch,
            `item "${item.label}" in section "${section.label}" is neither a valid link (href) nor a non-empty branch (items) — dead node`,
          ).toBe(true);
        }
      }
    });

    it("no item has both href and items — the group 1 short-circuit edge-case does not appear in production data", () => {
      for (const section of sidebar) {
        for (const item of walkItems(section.items)) {
          const hasBoth = item.href !== undefined && item.items !== undefined;
          expect(
            hasBoth,
            `item "${item.label}" in section "${section.label}" must not have both href and items`,
          ).toBe(false);
        }
      }
    });

    it("all leaf hrefs across the entire sidebar are unique", () => {
      const allLeaves = sidebar.flatMap((section) =>
        flattenItems(section.items),
      );
      const hrefs = allLeaves.map((leaf) => leaf.href);
      const uniqueHrefs = new Set(hrefs);
      expect(
        uniqueHrefs.size,
        `expected ${hrefs.length} unique hrefs but found ${uniqueHrefs.size} — duplicates: ${hrefs.filter((h, i) => hrefs.indexOf(h) !== i).join(", ")}`,
      ).toBe(hrefs.length);
    });
  });

  describe("group 3: dead-link integrity — every sidebar href resolves to a real page file", () => {
    it("route set contains more than 40 entries (sanity check for scan path)", () => {
      const pagesDir = join(import.meta.dirname, "../../pages");
      const entries = readdirSync(pagesDir, {
        recursive: true,
        encoding: "utf-8",
      });
      const routes = new Set(
        entries
          .filter((rel) => /\.(mdx|astro)$/.test(rel))
          .map((rel) => {
            const route = ("/" + rel.replace(/\.(mdx|astro)$/, "")).replace(
              /\/index$/,
              "",
            );
            return route || "/";
          }),
      );
      expect(routes.size).toBeGreaterThan(40);
    });

    it("every sidebar href resolves to an existing page file", () => {
      const pagesDir = join(import.meta.dirname, "../../pages");
      const entries = readdirSync(pagesDir, {
        recursive: true,
        encoding: "utf-8",
      });
      const routes = new Set(
        entries
          .filter((rel) => /\.(mdx|astro)$/.test(rel))
          .map((rel) => {
            const route = ("/" + rel.replace(/\.(mdx|astro)$/, "")).replace(
              /\/index$/,
              "",
            );
            return route || "/";
          }),
      );
      const allLeaves = sidebar.flatMap((section) =>
        flattenItems(section.items),
      );
      for (const leaf of allLeaves) {
        expect(
          routes.has(leaf.href),
          `sidebar href "${leaf.href}" (label: "${leaf.label}") does not match any page file`,
        ).toBe(true);
      }
    });
  });

  describe("group 4: orphan-page integrity — every page file appears in the sidebar or the allowlist", () => {
    function scanRoutes(): Set<string> {
      const pagesDir = join(import.meta.dirname, "../../pages");
      const entries = readdirSync(pagesDir, {
        recursive: true,
        encoding: "utf-8",
      });
      return new Set(
        entries
          .filter((rel) => /\.(mdx|astro)$/.test(rel))
          .map((rel) => {
            const route = ("/" + rel.replace(/\.(mdx|astro)$/, "")).replace(
              /\/index$/,
              "",
            );
            return route || "/";
          }),
      );
    }

    it("PAGES_NOT_IN_NAV is an array that contains '/'", () => {
      expect(Array.isArray(PAGES_NOT_IN_NAV)).toBe(true);
      expect(PAGES_NOT_IN_NAV).toContain("/");
    });

    it("every page file is either in the sidebar hrefs or in PAGES_NOT_IN_NAV — no orphaned pages", () => {
      const routes = scanRoutes();
      const sidebarHrefs = new Set(
        sidebar
          .flatMap((section) => flattenItems(section.items))
          .map((leaf) => leaf.href),
      );
      const allowlist = new Set<string>(PAGES_NOT_IN_NAV);

      for (const route of routes) {
        expect(
          sidebarHrefs.has(route) || allowlist.has(route),
          `page file route "${route}" is not listed in any sidebar section and is not in PAGES_NOT_IN_NAV — add it to the sidebar or to the allowlist`,
        ).toBe(true);
      }
    });
  });
});
