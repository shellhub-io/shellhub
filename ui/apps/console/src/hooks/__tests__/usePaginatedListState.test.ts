import { describe, it, expect } from "vitest";
import { act } from "@testing-library/react";
import {
  usePaginatedListState,
  type SortFieldDef,
} from "../usePaginatedListState";
import { renderHookWithRouter } from "@/tests/renderHookWithRouter";

type TestParams = {
  page: number;
  search: string;
};

const DEFAULTS: TestParams = {
  page: 1,
  search: "",
};

describe("usePaginatedListState — page", () => {
  it("reads page=1 from default URL", () => {
    const { result } = renderHookWithRouter(() =>
      usePaginatedListState({ defaults: DEFAULTS }),
    );
    expect(result.current.params.page).toBe(1);
  });

  it("reads page from URL search param", () => {
    const { result } = renderHookWithRouter(
      () => usePaginatedListState({ defaults: DEFAULTS }),
      { initialEntries: ["/?page=3"] },
    );
    expect(result.current.params.page).toBe(3);
  });

  it("setPage updates the page param", () => {
    const { result } = renderHookWithRouter(() =>
      usePaginatedListState({ defaults: DEFAULTS }),
    );

    act(() => {
      result.current.setPage(5);
    });

    expect(result.current.params.page).toBe(5);
  });
});

describe("usePaginatedListState — search", () => {
  it("reads empty search from default URL", () => {
    const { result } = renderHookWithRouter(() =>
      usePaginatedListState({ defaults: DEFAULTS }),
    );
    expect(result.current.params.search).toBe("");
  });

  it("reads search from URL search param", () => {
    const { result } = renderHookWithRouter(
      () => usePaginatedListState({ defaults: DEFAULTS }),
      { initialEntries: ["/?search=hello"] },
    );
    expect(result.current.params.search).toBe("hello");
  });

  it("setSearch updates the search param", () => {
    const { result } = renderHookWithRouter(() =>
      usePaginatedListState({ defaults: DEFAULTS }),
    );

    act(() => {
      result.current.setSearch("my-device");
    });

    expect(result.current.params.search).toBe("my-device");
  });
});

describe("usePaginatedListState — non-page setter resets page to 1", () => {
  it("setSearch resets page to 1 when page was > 1", () => {
    const { result } = renderHookWithRouter(
      () => usePaginatedListState({ defaults: DEFAULTS }),
      { initialEntries: ["/?page=4"] },
    );

    expect(result.current.params.page).toBe(4);

    act(() => {
      result.current.setSearch("filter");
    });

    expect(result.current.params.page).toBe(1);
  });
});

describe("usePaginatedListState — reset", () => {
  it("reset() clears all params back to defaults", () => {
    const { result } = renderHookWithRouter(
      () => usePaginatedListState({ defaults: DEFAULTS }),
      { initialEntries: ["/?page=3&search=filter"] },
    );

    expect(result.current.params.page).toBe(3);
    expect(result.current.params.search).toBe("filter");

    act(() => {
      result.current.reset();
    });

    expect(result.current.params.page).toBe(1);
    expect(result.current.params.search).toBe("");
  });
});

describe("usePaginatedListState — replace-history mode omits default-valued params", () => {
  it("does not write page=1 to URL when it equals the default", () => {
    const { result } = renderHookWithRouter(() =>
      usePaginatedListState({ defaults: DEFAULTS }),
    );

    act(() => {
      result.current.setPage(2);
    });
    act(() => {
      result.current.setPage(1);
    });

    expect(result.current.params.page).toBe(1);

    expect(result.current.searchString).not.toContain("page=1");
  });

  it("does not write search= when search equals the default (empty string)", () => {
    const { result } = renderHookWithRouter(() =>
      usePaginatedListState({ defaults: DEFAULTS }),
    );

    act(() => {
      result.current.setSearch("tmp");
    });
    act(() => {
      result.current.setSearch("");
    });

    expect(result.current.params.search).toBe("");
    expect(result.current.searchString).not.toContain("search=");
  });
});

describe("usePaginatedListState — unrelated params are preserved", () => {
  it("leaves unrelated URL params untouched when updating page", () => {
    const { result } = renderHookWithRouter(
      () => usePaginatedListState({ defaults: DEFAULTS }),
      { initialEntries: ["/?unrelated=keep&page=2"] },
    );

    act(() => {
      result.current.setPage(3);
    });

    expect(result.current.searchString).toContain("unrelated=keep");
    expect(result.current.params.page).toBe(3);
  });

  it("leaves unrelated URL params untouched when resetting", () => {
    const { result } = renderHookWithRouter(
      () => usePaginatedListState({ defaults: DEFAULTS }),
      { initialEntries: ["/?unrelated=keep&page=2&search=foo"] },
    );

    act(() => {
      result.current.reset();
    });

    expect(result.current.searchString).toContain("unrelated=keep");
    expect(result.current.params.page).toBe(1);
    expect(result.current.params.search).toBe("");
  });
});

type SortableParams = {
  page: number;
  search: string;
  sortField: string;
  sortOrder: "asc" | "desc";
};

const SORT_DEFAULTS: SortableParams = {
  page: 1,
  search: "",
  sortField: "name",
  sortOrder: "asc",
};

const SORT_FIELDS: SortFieldDef[] = [
  { field: "name", initialOrder: "asc" },
  { field: "created_at", initialOrder: "desc" },
];

describe("usePaginatedListState — handleSort", () => {
  it("toggles order from asc to desc when clicking the same field", () => {
    const { result } = renderHookWithRouter(() =>
      usePaginatedListState({
        defaults: SORT_DEFAULTS,
        sortFields: SORT_FIELDS,
      }),
    );

    act(() => {
      result.current.handleSort("name");
    });

    expect(result.current.params.sortField).toBe("name");
    expect(result.current.params.sortOrder).toBe("desc");
  });

  it("toggles order from desc to asc when clicking the same field again", () => {
    const { result } = renderHookWithRouter(
      () =>
        usePaginatedListState({
          defaults: SORT_DEFAULTS,
          sortFields: SORT_FIELDS,
        }),
      { initialEntries: ["/?sortField=name&sortOrder=desc"] },
    );

    act(() => {
      result.current.handleSort("name");
    });

    expect(result.current.params.sortField).toBe("name");
    expect(result.current.params.sortOrder).toBe("asc");
  });

  it("uses the field's initialOrder when switching to a different field", () => {
    const { result } = renderHookWithRouter(() =>
      usePaginatedListState({
        defaults: SORT_DEFAULTS,
        sortFields: SORT_FIELDS,
      }),
    );

    act(() => {
      result.current.handleSort("created_at");
    });

    expect(result.current.params.sortField).toBe("created_at");
    expect(result.current.params.sortOrder).toBe("desc");
  });

  it("resets page to 1 when sorting changes", () => {
    const { result } = renderHookWithRouter(
      () =>
        usePaginatedListState({
          defaults: SORT_DEFAULTS,
          sortFields: SORT_FIELDS,
        }),
      { initialEntries: ["/?page=3"] },
    );

    expect(result.current.params.page).toBe(3);

    act(() => {
      result.current.handleSort("name");
    });

    expect(result.current.params.page).toBe(1);
  });

  it("resets page to 1 when switching to a different sort field", () => {
    const { result } = renderHookWithRouter(
      () =>
        usePaginatedListState({
          defaults: SORT_DEFAULTS,
          sortFields: SORT_FIELDS,
        }),
      { initialEntries: ["/?page=5&sortField=name&sortOrder=asc"] },
    );

    act(() => {
      result.current.handleSort("created_at");
    });

    expect(result.current.params.page).toBe(1);
    expect(result.current.params.sortField).toBe("created_at");
    expect(result.current.params.sortOrder).toBe("desc");
  });
});

type FilterParams = {
  page: number;
  search: string;
  status: string;
};

const VALID_STATUSES = ["active", "inactive", "pending"] as const;

const FILTER_DEFAULTS: FilterParams = {
  page: 1,
  search: "",
  status: "active",
};

describe("usePaginatedListState — scalar filter with allowlist", () => {
  it("setFilter updates a scalar param and resets page", () => {
    const { result } = renderHookWithRouter(
      () =>
        usePaginatedListState({
          defaults: FILTER_DEFAULTS,
          constraints: { status: VALID_STATUSES },
        }),
      { initialEntries: ["/?page=3"] },
    );

    act(() => {
      result.current.setFilter("status", "inactive");
    });

    expect(result.current.params.status).toBe("inactive");
    expect(result.current.params.page).toBe(1);
  });

  it("rejects values not in the allowlist and falls back to default", () => {
    const { result } = renderHookWithRouter(
      () =>
        usePaginatedListState({
          defaults: FILTER_DEFAULTS,
          constraints: { status: VALID_STATUSES },
        }),
      { initialEntries: ["/?status=inactive"] },
    );

    expect(result.current.params.status).toBe("inactive");

    act(() => {
      result.current.setFilter("status", "unknown");
    });

    expect(result.current.params.status).toBe("active");
  });

  it("omits the param from the URL when it equals the default", () => {
    const { result } = renderHookWithRouter(
      () =>
        usePaginatedListState({
          defaults: FILTER_DEFAULTS,
          constraints: { status: VALID_STATUSES },
        }),
      { initialEntries: ["/?status=inactive"] },
    );

    act(() => {
      result.current.setFilter("status", "active");
    });

    expect(result.current.params.status).toBe("active");
    expect(result.current.searchString).not.toContain("status=");
  });
});

type ArrayFilterParams = {
  page: number;
  search: string;
  tags: string[];
};

const VALID_TAGS = ["web", "prod", "staging", "api"] as const;

const ARRAY_FILTER_DEFAULTS: ArrayFilterParams = {
  page: 1,
  search: "",
  tags: [],
};

describe("usePaginatedListState — array filters", () => {
  it("setArrayFilter writes repeated-key params to the URL", () => {
    const { result } = renderHookWithRouter(() =>
      usePaginatedListState({
        defaults: ARRAY_FILTER_DEFAULTS,
        constraints: { tags: VALID_TAGS },
      }),
    );

    act(() => {
      result.current.setArrayFilter("tags", ["web", "prod"]);
    });

    expect(result.current.params.tags).toEqual(["web", "prod"]);
    const sp = new URLSearchParams(result.current.searchString);
    expect(sp.getAll("tags")).toEqual(["web", "prod"]);
  });

  it("filters out invalid values against the allowlist", () => {
    const { result } = renderHookWithRouter(() =>
      usePaginatedListState({
        defaults: ARRAY_FILTER_DEFAULTS,
        constraints: { tags: VALID_TAGS },
      }),
    );

    act(() => {
      result.current.setArrayFilter("tags", ["web", "invalid", "prod"]);
    });

    expect(result.current.params.tags).toEqual(["web", "prod"]);
  });

  it("omits tag params from URL when array is reset to default (empty)", () => {
    const { result } = renderHookWithRouter(
      () =>
        usePaginatedListState({
          defaults: ARRAY_FILTER_DEFAULTS,
          constraints: { tags: VALID_TAGS },
        }),
      { initialEntries: ["/?tags=web&tags=prod"] },
    );

    act(() => {
      result.current.setArrayFilter("tags", []);
    });

    expect(result.current.params.tags).toEqual([]);
    expect(result.current.searchString).not.toContain("tags=");
  });

  it("resets page to 1 when array filter changes", () => {
    const { result } = renderHookWithRouter(
      () =>
        usePaginatedListState({
          defaults: ARRAY_FILTER_DEFAULTS,
          constraints: { tags: VALID_TAGS },
        }),
      { initialEntries: ["/?page=4"] },
    );

    act(() => {
      result.current.setArrayFilter("tags", ["api"]);
    });

    expect(result.current.params.page).toBe(1);
    expect(result.current.params.tags).toEqual(["api"]);
  });

  it("round-trips: values written via setArrayFilter are read back via getAll", () => {
    const { result } = renderHookWithRouter(() =>
      usePaginatedListState({
        defaults: ARRAY_FILTER_DEFAULTS,
      }),
    );

    act(() => {
      result.current.setArrayFilter("tags", ["staging", "api"]);
    });

    const sp = new URLSearchParams(result.current.searchString);
    expect(sp.getAll("tags")).toEqual(["staging", "api"]);
    expect(result.current.params.tags).toEqual(["staging", "api"]);
  });
});

describe("usePaginatedListState — mapArrayFilter", () => {
  it("applies the functional updater to the current array value", () => {
    const { result } = renderHookWithRouter(
      () =>
        usePaginatedListState({
          defaults: ARRAY_FILTER_DEFAULTS,
          constraints: { tags: VALID_TAGS },
        }),
      { initialEntries: ["/?tags=web&tags=prod"] },
    );

    act(() => {
      result.current.mapArrayFilter("tags", (tags) =>
        tags.filter((t) => t !== "web"),
      );
    });

    expect(result.current.params.tags).toEqual(["prod"]);
    const sp = new URLSearchParams(result.current.searchString);
    expect(sp.getAll("tags")).toEqual(["prod"]);
  });

  it("resets page to 1 when the array changes", () => {
    const { result } = renderHookWithRouter(
      () =>
        usePaginatedListState({
          defaults: ARRAY_FILTER_DEFAULTS,
          constraints: { tags: VALID_TAGS },
        }),
      { initialEntries: ["/?page=4&tags=web"] },
    );

    act(() => {
      result.current.mapArrayFilter("tags", (tags) => [...tags, "prod"]);
    });

    expect(result.current.params.page).toBe(1);
    expect(result.current.params.tags).toEqual(["web", "prod"]);
  });
});

type PrefixedParams = {
  page: number;
  search: string;
};

const PREFIXED_DEFAULTS: PrefixedParams = { page: 1, search: "" };

describe("usePaginatedListState — prefix option", () => {
  it("namespaces params with the given prefix", () => {
    const { result } = renderHookWithRouter(() =>
      usePaginatedListState({ defaults: PREFIXED_DEFAULTS, prefix: "devices" }),
    );

    act(() => {
      result.current.setPage(3);
    });

    expect(result.current.params.page).toBe(3);
    const sp = new URLSearchParams(result.current.searchString);
    expect(sp.get("devices.page")).toBe("3");
    expect(sp.get("page")).toBeNull();
  });

  it("reads params from prefixed URL keys", () => {
    const { result } = renderHookWithRouter(
      () =>
        usePaginatedListState({
          defaults: PREFIXED_DEFAULTS,
          prefix: "devices",
        }),
      { initialEntries: ["/?devices.page=5&devices.search=hello"] },
    );

    expect(result.current.params.page).toBe(5);
    expect(result.current.params.search).toBe("hello");
  });

  it("reads page from its own prefixed key when the URL contains both prefixed keys", () => {
    const { result: r1 } = renderHookWithRouter(
      () =>
        usePaginatedListState({
          defaults: PREFIXED_DEFAULTS,
          prefix: "devices",
        }),
      { initialEntries: ["/?devices.page=2&sessions.page=7"] },
    );
    const { result: r2 } = renderHookWithRouter(
      () =>
        usePaginatedListState({
          defaults: PREFIXED_DEFAULTS,
          prefix: "sessions",
        }),
      { initialEntries: ["/?devices.page=2&sessions.page=7"] },
    );

    expect(r1.current.params.page).toBe(2);
    expect(r2.current.params.page).toBe(7);
  });

  it("updating one prefixed instance does not clobber another instance's params", () => {
    function useTwo() {
      const h1 = usePaginatedListState({
        defaults: PREFIXED_DEFAULTS,
        prefix: "devices",
      });
      const h2 = usePaginatedListState({
        defaults: PREFIXED_DEFAULTS,
        prefix: "sessions",
      });
      return { h1, h2 };
    }

    const { result } = renderHookWithRouter(useTwo, {
      initialEntries: ["/?devices.page=2&sessions.page=7"],
    });

    act(() => {
      result.current.h1.setPage(5);
    });

    expect(result.current.h1.params.page).toBe(5);
    expect(result.current.h2.params.page).toBe(7);
  });

  it("does not touch the unprefixed key when a prefix is set", () => {
    const { result } = renderHookWithRouter(
      () => usePaginatedListState({ defaults: PREFIXED_DEFAULTS, prefix: "a" }),
      { initialEntries: ["/?page=99&a.page=2"] },
    );

    expect(result.current.params.page).toBe(2);
  });

  it("preserves prefixed params when updating another prefixed instance", () => {
    function useTwo() {
      const h1 = usePaginatedListState({
        defaults: PREFIXED_DEFAULTS,
        prefix: "devices",
      });
      const h2 = usePaginatedListState({
        defaults: PREFIXED_DEFAULTS,
        prefix: "sessions",
      });
      return { h1, h2 };
    }

    const { result } = renderHookWithRouter(useTwo, {
      initialEntries: ["/?devices.page=1&sessions.page=3"],
    });

    act(() => {
      result.current.h1.setPage(5);
    });

    expect(result.current.h1.params.page).toBe(5);
    expect(result.current.h2.params.page).toBe(3);
  });
});
