import { describe, it, expect } from "vitest";
import {
  parseListParams,
  serializeListParams,
  type ListParamDefaults,
} from "../paginatedListParams";

// ── Shared fixture ────────────────────────────────────────────────────────────

const VALID_STATUSES = ["accepted", "pending", "rejected"] as const;
const PER_PAGE_ALLOW = [10, 25, 50] as const;

type Status = (typeof VALID_STATUSES)[number];

type TestParams = {
  page: number;
  perPage: number;
  status: Status;
  tags: string[];
};

const DEFAULTS: ListParamDefaults<TestParams> = {
  page: 1,
  perPage: 10,
  status: "accepted",
  tags: [],
};

function params(init?: Record<string, string | string[]>): URLSearchParams {
  const p = new URLSearchParams();
  if (init) {
    for (const [k, v] of Object.entries(init)) {
      if (Array.isArray(v)) {
        for (const item of v) p.append(k, item);
      } else {
        p.set(k, v);
      }
    }
  }
  return p;
}

// ── parseListParams ───────────────────────────────────────────────────────────

describe("parseListParams", () => {
  describe("page", () => {
    it("returns default page when the param is absent", () => {
      const result = parseListParams<TestParams>(params(), DEFAULTS, {});
      expect(result.page).toBe(1);
    });

    it("parses a valid page integer", () => {
      const result = parseListParams<TestParams>(params({ page: "3" }), DEFAULTS, {});
      expect(result.page).toBe(3);
    });

    it("falls back to default when page is 0", () => {
      const result = parseListParams<TestParams>(params({ page: "0" }), DEFAULTS, {});
      expect(result.page).toBe(1);
    });

    it("falls back to default when page is negative", () => {
      const result = parseListParams<TestParams>(params({ page: "-1" }), DEFAULTS, {});
      expect(result.page).toBe(1);
    });

    it("falls back to default when page is non-numeric", () => {
      const result = parseListParams<TestParams>(params({ page: "abc" }), DEFAULTS, {});
      expect(result.page).toBe(1);
    });

    it("falls back to default when page is a float", () => {
      const result = parseListParams<TestParams>(params({ page: "2.5" }), DEFAULTS, {});
      expect(result.page).toBe(1);
    });
  });

  describe("perPage", () => {
    it("returns default perPage when the param is absent", () => {
      const result = parseListParams<TestParams>(params(), DEFAULTS, {
        perPage: PER_PAGE_ALLOW,
      });
      expect(result.perPage).toBe(10);
    });

    it("parses a valid perPage from the allowlist", () => {
      const result = parseListParams<TestParams>(params({ perPage: "25" }), DEFAULTS, {
        perPage: PER_PAGE_ALLOW,
      });
      expect(result.perPage).toBe(25);
    });

    it("falls back to default when perPage is not in the allowlist", () => {
      const result = parseListParams<TestParams>(params({ perPage: "99" }), DEFAULTS, {
        perPage: PER_PAGE_ALLOW,
      });
      expect(result.perPage).toBe(10);
    });

    it("falls back to default when perPage is non-numeric", () => {
      const result = parseListParams<TestParams>(params({ perPage: "many" }), DEFAULTS, {
        perPage: PER_PAGE_ALLOW,
      });
      expect(result.perPage).toBe(10);
    });

    it("returns the raw numeric value when no allowlist is provided", () => {
      const result = parseListParams<TestParams>(params({ perPage: "50" }), DEFAULTS, {});
      expect(result.perPage).toBe(50);
    });

    it("falls back to default when perPage is 0 and no allowlist", () => {
      const result = parseListParams<TestParams>(params({ perPage: "0" }), DEFAULTS, {});
      expect(result.perPage).toBe(10);
    });
  });

  describe("enum-typed dimension (status)", () => {
    it("returns default status when the param is absent", () => {
      const result = parseListParams<TestParams>(params(), DEFAULTS, {
        status: VALID_STATUSES,
      });
      expect(result.status).toBe("accepted");
    });

    it("parses a valid status value", () => {
      const result = parseListParams<TestParams>(
        params({ status: "pending" }),
        DEFAULTS,
        { status: VALID_STATUSES },
      );
      expect(result.status).toBe("pending");
    });

    it("falls back to default for an invalid status", () => {
      const result = parseListParams<TestParams>(
        params({ status: "unknown" }),
        DEFAULTS,
        { status: VALID_STATUSES },
      );
      expect(result.status).toBe("accepted");
    });

    it("returns the raw value when no valid array is provided for status", () => {
      const result = parseListParams<TestParams>(
        params({ status: "anything" }),
        DEFAULTS,
        {},
      );
      expect(result.status).toBe("anything");
    });
  });

  describe("array dimension (tags)", () => {
    it("returns default empty array when no tags param is present", () => {
      const result = parseListParams<TestParams>(params(), DEFAULTS, {});
      expect(result.tags).toEqual([]);
    });

    it("collects repeated tag params into an array", () => {
      const result = parseListParams<TestParams>(
        params({ tags: ["web", "prod"] }),
        DEFAULTS,
        {},
      );
      expect(result.tags).toEqual(["web", "prod"]);
    });

    it("returns single tag as a one-element array", () => {
      const result = parseListParams<TestParams>(
        params({ tags: ["api"] }),
        DEFAULTS,
        {},
      );
      expect(result.tags).toEqual(["api"]);
    });

    it("filters tags against a valid array when provided", () => {
      const result = parseListParams<TestParams>(
        params({ tags: ["web", "unknown", "prod"] }),
        DEFAULTS,
        { tags: ["web", "prod", "staging"] },
      );
      expect(result.tags).toEqual(["web", "prod"]);
    });

    it("returns empty array when all tags are invalid against the allowlist", () => {
      const result = parseListParams<TestParams>(
        params({ tags: ["bad", "nope"] }),
        DEFAULTS,
        { tags: ["web", "prod"] },
      );
      expect(result.tags).toEqual([]);
    });

    it("returns the same array reference when content is unchanged", () => {
      const sp = params({ tags: ["web", "prod"] });
      const first = parseListParams<TestParams>(sp, DEFAULTS, {}, undefined);
      const second = parseListParams<TestParams>(sp, DEFAULTS, {}, first);
      expect(second.tags).toBe(first.tags);
    });

    it("returns the same array reference when tags are the same but in different order", () => {
      const sp1 = params({ tags: ["web", "prod"] });
      const sp2 = params({ tags: ["prod", "web"] });
      const first = parseListParams<TestParams>(sp1, DEFAULTS, {}, undefined);
      const second = parseListParams<TestParams>(sp2, DEFAULTS, {}, first);
      expect(second.tags).toBe(first.tags);
    });

    it("returns a new array reference when content changes", () => {
      const sp1 = params({ tags: ["web", "prod"] });
      const sp2 = params({ tags: ["web", "staging"] });
      const first = parseListParams<TestParams>(sp1, DEFAULTS, {}, undefined);
      const second = parseListParams<TestParams>(sp2, DEFAULTS, {}, first);
      expect(second.tags).not.toBe(first.tags);
      expect(second.tags).toEqual(["web", "staging"]);
    });
  });
});

// ── serializeListParams ───────────────────────────────────────────────────────

describe("serializeListParams", () => {
  describe("page", () => {
    it("omits page when equal to default", () => {
      const sp = serializeListParams<TestParams>({ page: 1, perPage: 10, status: "accepted", tags: [] }, DEFAULTS);
      expect(sp.get("page")).toBeNull();
    });

    it("writes page when different from default", () => {
      const sp = serializeListParams<TestParams>({ page: 3, perPage: 10, status: "accepted", tags: [] }, DEFAULTS);
      expect(sp.get("page")).toBe("3");
    });
  });

  describe("perPage", () => {
    it("omits perPage when equal to default", () => {
      const sp = serializeListParams<TestParams>({ page: 1, perPage: 10, status: "accepted", tags: [] }, DEFAULTS);
      expect(sp.get("perPage")).toBeNull();
    });

    it("writes perPage when different from default", () => {
      const sp = serializeListParams<TestParams>({ page: 1, perPage: 25, status: "accepted", tags: [] }, DEFAULTS);
      expect(sp.get("perPage")).toBe("25");
    });
  });

  describe("enum dimension (status)", () => {
    it("omits status when equal to default", () => {
      const sp = serializeListParams<TestParams>({ page: 1, perPage: 10, status: "accepted", tags: [] }, DEFAULTS);
      expect(sp.get("status")).toBeNull();
    });

    it("writes status when different from default", () => {
      const sp = serializeListParams<TestParams>({ page: 1, perPage: 10, status: "pending", tags: [] }, DEFAULTS);
      expect(sp.get("status")).toBe("pending");
    });
  });

  describe("array dimension (tags)", () => {
    it("omits tags when equal to default (empty)", () => {
      const sp = serializeListParams<TestParams>({ page: 1, perPage: 10, status: "accepted", tags: [] }, DEFAULTS);
      expect(sp.getAll("tags")).toEqual([]);
    });

    it("writes repeated-key entries for each tag", () => {
      const sp = serializeListParams<TestParams>(
        { page: 1, perPage: 10, status: "accepted", tags: ["web", "prod"] },
        DEFAULTS,
      );
      expect(sp.getAll("tags")).toEqual(["web", "prod"]);
    });

    it("omits tags when the array matches the default exactly", () => {
      const defaultsWithTags: ListParamDefaults<TestParams> = { ...DEFAULTS, tags: ["web"] };
      const sp = serializeListParams<TestParams>(
        { page: 1, perPage: 10, status: "accepted", tags: ["web"] },
        defaultsWithTags,
      );
      expect(sp.getAll("tags")).toEqual([]);
    });
  });

  describe("combined", () => {
    it("produces an empty URLSearchParams when all params equal defaults", () => {
      const sp = serializeListParams<TestParams>(
        { page: 1, perPage: 10, status: "accepted", tags: [] },
        DEFAULTS,
      );
      expect(sp.toString()).toBe("");
    });

    it("serializes only the non-default dimensions", () => {
      const sp = serializeListParams<TestParams>(
        { page: 2, perPage: 25, status: "accepted", tags: ["web"] },
        DEFAULTS,
      );
      expect(sp.get("page")).toBe("2");
      expect(sp.get("perPage")).toBe("25");
      expect(sp.get("status")).toBeNull();
      expect(sp.getAll("tags")).toEqual(["web"]);
    });
  });
});
