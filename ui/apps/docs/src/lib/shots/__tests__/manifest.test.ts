import { describe, expect, it } from "vitest";
import type { ShotDeclaration } from "@/lib/shots/manifest";
import { buildManifest } from "@/lib/shots/manifest";

function declare(overrides: Partial<ShotDeclaration> = {}): ShotDeclaration {
  return {
    id: "device-list",
    route: "/devices",
    page: "/getting-started/quick-start/",
    ...overrides,
  };
}

describe("buildManifest", () => {
  describe("deduplication", () => {
    it("collapses the same shot declared on two pages into one entry", () => {
      const manifest = buildManifest([
        declare({ page: "/getting-started/quick-start/" }),
        declare({ page: "/guides/devices/" }),
      ]);

      expect(manifest.shots).toHaveLength(1);
      expect(manifest.shots[0].usedBy).toEqual([
        "/getting-started/quick-start/",
        "/guides/devices/",
      ]);
    });

    it("records a page once when the same shot appears twice on it", () => {
      const manifest = buildManifest([
        declare({ page: "/guides/devices/" }),
        declare({ page: "/guides/devices/" }),
      ]);

      expect(manifest.shots[0].usedBy).toEqual(["/guides/devices/"]);
    });

    it("keeps distinct ids as separate shots", () => {
      const manifest = buildManifest([
        declare({ id: "device-list" }),
        declare({ id: "session-list", route: "/sessions" }),
      ]);

      expect(manifest.shots.map((shot) => shot.id)).toEqual([
        "device-list",
        "session-list",
      ]);
    });

    it("rejects one id declaring two different captures", () => {
      const entries = [
        declare({ id: "device-list", route: "/devices" }),
        declare({ id: "device-list", route: "/sessions" }),
      ];

      expect(() => buildManifest(entries)).toThrow(/device-list/);
    });
  });

  describe("defaults", () => {
    it("falls back to a 1440x900 viewport and the community edition", () => {
      const manifest = buildManifest([declare()]);

      expect(manifest.shots[0]).toMatchObject({
        viewport: { width: 1440, height: 900 },
        edition: "ce",
      });
    });

    it("keeps explicit values over the defaults", () => {
      const manifest = buildManifest([
        declare({ viewport: { width: 800, height: 600 }, edition: "enterprise" }),
      ]);

      expect(manifest.shots[0]).toMatchObject({
        viewport: { width: 800, height: 600 },
        edition: "enterprise",
      });
    });

    it("preserves the capture coordinates the author supplied", () => {
      const of = { role: "table" };
      const before = [{ click: { role: "button", name: "Sort by Hostname" } }];

      const manifest = buildManifest([declare({ of, before })]);

      expect(manifest.shots[0]).toMatchObject({ of, before });
    });

    it("omits optional coordinates that were never declared", () => {
      const manifest = buildManifest([declare()]);

      expect(manifest.shots[0]).not.toHaveProperty("of");
      expect(manifest.shots[0]).not.toHaveProperty("before");
    });
  });

  describe("stability", () => {
    it("sorts shots by id regardless of the order pages rendered in", () => {
      const ids = ["session-list", "dashboard", "device-list"];

      const manifest = buildManifest(
        ids.map((id) => declare({ id, route: `/${id}` })),
      );

      expect(manifest.shots.map((shot) => shot.id)).toEqual([
        "dashboard",
        "device-list",
        "session-list",
      ]);
    });

    it("sorts usedBy so a reordered build produces the same manifest", () => {
      const forward = buildManifest([
        declare({ page: "/a/" }),
        declare({ page: "/b/" }),
      ]);
      const reversed = buildManifest([
        declare({ page: "/b/" }),
        declare({ page: "/a/" }),
      ]);

      expect(reversed).toEqual(forward);
      expect(forward.shots[0].usedBy).toEqual(["/a/", "/b/"]);
    });

    it("returns an empty shot list when no page declared one", () => {
      expect(buildManifest([]).shots).toEqual([]);
    });
  });
});
