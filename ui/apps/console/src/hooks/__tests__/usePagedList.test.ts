import { describe, it, expect, afterEach, vi } from "vitest";
import { act, renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { useLocation } from "react-router-dom";
import { server, jsonWithTotal } from "@/tests/msw";
import { paginatedQueryOptions } from "@/api/pagination";
import { getPublicKeys, getPublicKeysQueryKey } from "@/client";
import { PER_PAGE } from "@/utils/pagination";
import { renderHookWithRouter } from "@/tests/renderHookWithRouter";
import { createTestQueryClient, createTestWrapper } from "@/tests/wrapper";
import { usePagedList, type ListStateConfig } from "../usePagedList";

function pendingQuery() {
  return {
    queryKey: ["pending"],
    queryFn: () => new Promise<{ data: never[]; totalCount: number }>(() => {}),
  };
}

function useListState<T extends Record<string, unknown> & { page: number }>(
  config: ListStateConfig<T>,
) {
  const searchString = new URLSearchParams(useLocation().search).toString();
  return { ...usePagedList({ ...config, query: pendingQuery }), searchString };
}

type TestParams = {
  page: number;
  search: string;
};

const DEFAULTS: TestParams = {
  page: 1,
  search: "",
};

describe("usePagedList — page", () => {
  it("setPage updates the page param", () => {
    const { result } = renderHookWithRouter(() =>
      useListState({ defaults: DEFAULTS }),
    );

    act(() => {
      result.current.setPage(5);
    });

    expect(result.current.params.page).toBe(5);
  });
});

describe("usePagedList — search", () => {
  it("reads empty search from default URL", () => {
    const { result } = renderHookWithRouter(() =>
      useListState({ defaults: DEFAULTS }),
    );
    expect(result.current.params.search).toBe("");
  });

  it("reads search from URL search param", () => {
    const { result } = renderHookWithRouter(
      () => useListState({ defaults: DEFAULTS }),
      { initialEntries: ["/?search=hello"] },
    );
    expect(result.current.params.search).toBe("hello");
  });

  it("setSearch updates the search param", () => {
    const { result } = renderHookWithRouter(() =>
      useListState({ defaults: DEFAULTS }),
    );

    act(() => {
      result.current.setSearch("my-device");
    });

    expect(result.current.params.search).toBe("my-device");
  });
});

describe("usePagedList — non-page setter resets page to 1", () => {
  it("setSearch resets page to 1 when page was > 1", () => {
    const { result } = renderHookWithRouter(
      () => useListState({ defaults: DEFAULTS }),
      { initialEntries: ["/?page=4"] },
    );

    expect(result.current.params.page).toBe(4);

    act(() => {
      result.current.setSearch("filter");
    });

    expect(result.current.params.page).toBe(1);
  });
});

describe("usePagedList — replace-history mode omits default-valued params", () => {
  it("does not write page=1 to URL when it equals the default", () => {
    const { result } = renderHookWithRouter(() =>
      useListState({ defaults: DEFAULTS }),
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
      useListState({ defaults: DEFAULTS }),
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

describe("usePagedList — unrelated params are preserved", () => {
  it("leaves unrelated URL params untouched when updating page", () => {
    const { result } = renderHookWithRouter(
      () => useListState({ defaults: DEFAULTS }),
      { initialEntries: ["/?unrelated=keep&page=2"] },
    );

    act(() => {
      result.current.setPage(3);
    });

    expect(result.current.searchString).toContain("unrelated=keep");
    expect(result.current.params.page).toBe(3);
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

describe("usePagedList — scalar filter with allowlist", () => {
  it("setFilter updates a scalar param and resets page", () => {
    const { result } = renderHookWithRouter(
      () =>
        useListState({
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
        useListState({
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
        useListState({
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

describe("usePagedList — array filters", () => {
  it("writes an array filter to the URL as repeated keys", () => {
    const { result } = renderHookWithRouter(() =>
      useListState({
        defaults: ARRAY_FILTER_DEFAULTS,
        constraints: { tags: VALID_TAGS },
      }),
    );

    act(() => {
      result.current.setFilter("tags", ["web", "prod"]);
    });

    expect(result.current.params.tags).toEqual(["web", "prod"]);
    const sp = new URLSearchParams(result.current.searchString);
    expect(sp.getAll("tags")).toEqual(["web", "prod"]);
  });

  it("filters out invalid values against the allowlist", () => {
    const { result } = renderHookWithRouter(() =>
      useListState({
        defaults: ARRAY_FILTER_DEFAULTS,
        constraints: { tags: VALID_TAGS },
      }),
    );

    act(() => {
      result.current.setFilter("tags", ["web", "invalid", "prod"]);
    });

    expect(result.current.params.tags).toEqual(["web", "prod"]);
  });

  it("omits tag params from URL when array is reset to default (empty)", () => {
    const { result } = renderHookWithRouter(
      () =>
        useListState({
          defaults: ARRAY_FILTER_DEFAULTS,
          constraints: { tags: VALID_TAGS },
        }),
      { initialEntries: ["/?tags=web&tags=prod"] },
    );

    act(() => {
      result.current.setFilter("tags", []);
    });

    expect(result.current.params.tags).toEqual([]);
    expect(result.current.searchString).not.toContain("tags=");
  });

  it("resets page to 1 when array filter changes", () => {
    const { result } = renderHookWithRouter(
      () =>
        useListState({
          defaults: ARRAY_FILTER_DEFAULTS,
          constraints: { tags: VALID_TAGS },
        }),
      { initialEntries: ["/?page=4"] },
    );

    act(() => {
      result.current.setFilter("tags", ["api"]);
    });

    expect(result.current.params.page).toBe(1);
    expect(result.current.params.tags).toEqual(["api"]);
  });
});

describe("usePagedList — mapArrayFilter", () => {
  it("applies the functional updater to the current array value", () => {
    const { result } = renderHookWithRouter(
      () =>
        useListState({
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
        useListState({
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

describe("usePagedList — prefix option", () => {
  it("namespaces params with the given prefix", () => {
    const { result } = renderHookWithRouter(() =>
      useListState({ defaults: PREFIXED_DEFAULTS, prefix: "devices" }),
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
        useListState({
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
        useListState({
          defaults: PREFIXED_DEFAULTS,
          prefix: "devices",
        }),
      { initialEntries: ["/?devices.page=2&sessions.page=7"] },
    );
    const { result: r2 } = renderHookWithRouter(
      () =>
        useListState({
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
      const h1 = useListState({
        defaults: PREFIXED_DEFAULTS,
        prefix: "devices",
      });
      const h2 = useListState({
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
      () => useListState({ defaults: PREFIXED_DEFAULTS, prefix: "a" }),
      { initialEntries: ["/?page=99&a.page=2"] },
    );

    expect(result.current.params.page).toBe(2);
  });

  it("preserves prefixed params when updating another prefixed instance", () => {
    function useTwo() {
      const h1 = useListState({
        defaults: PREFIXED_DEFAULTS,
        prefix: "devices",
      });
      const h2 = useListState({
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

const PUBLIC_KEYS = "*/api/sshkeys/public-keys";

function servePublicKeys(totalCount: number) {
  const filters: (string | null)[] = [];
  server.use(
    http.get(PUBLIC_KEYS, ({ request }) => {
      filters.push(new URL(request.url).searchParams.get("filter"));
      return jsonWithTotal([], totalCount);
    }),
  );
  return filters;
}

function usePublicKeyList() {
  return usePagedList({
    defaults: DEFAULTS,
    query: (params) =>
      paginatedQueryOptions(getPublicKeys, getPublicKeysQueryKey, {
        query: {
          page: params.page,
          per_page: PER_PAGE,
          ...(params.search ? { filter: params.search } : {}),
        },
      }),
  });
}

describe("usePagedList — query", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("sends a typed search only after the debounce", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const filters = servePublicKeys(0);
    const { result } = renderHookWithRouter(usePublicKeyList);
    await waitFor(() => expect(result.current.isLoading).toBe(false));

    act(() => {
      result.current.setSearch("alice");
    });

    act(() => {
      vi.advanceTimersByTime(250);
    });

    expect(result.current.params.search).toBe("alice");
    expect(result.current.appliedSearch).toBe("");
    expect(filters).not.toContain("alice");

    act(() => {
      vi.advanceTimersByTime(100);
    });

    await waitFor(() => expect(filters).toContain("alice"));
    expect(result.current.appliedSearch).toBe("alice");
  });

  it("moves to the last page once the total shrinks below the current page", async () => {
    servePublicKeys(PER_PAGE + 1);
    const { result } = renderHookWithRouter(usePublicKeyList, {
      initialEntries: ["/?page=3"],
    });

    expect(result.current.params.page).toBe(3);
    await waitFor(() => expect(result.current.params.page).toBe(2));
    expect(result.current.totalPages).toBe(2);
  });

  it("waits for a refetch before moving off a page the cache counts out", async () => {
    servePublicKeys(2 * PER_PAGE + 1);
    const queryClient = createTestQueryClient();
    queryClient.setQueryData(
      getPublicKeysQueryKey({ query: { page: 3, per_page: PER_PAGE } }),
      { data: [], totalCount: 2 * PER_PAGE },
    );
    const { result } = renderHook(usePublicKeyList, {
      wrapper: createTestWrapper({ queryClient, initialEntries: ["/?page=3"] }),
    });

    await waitFor(() => expect(result.current.totalPages).toBe(3));
    expect(result.current.params.page).toBe(3);
  });

  it("moves to the first page when the list is empty", async () => {
    servePublicKeys(0);
    const { result } = renderHookWithRouter(usePublicKeyList, {
      initialEntries: ["/?page=2"],
    });

    await waitFor(() => expect(result.current.params.page).toBe(1));
  });

  it("surfaces a failed request as error", async () => {
    server.use(
      http.get(PUBLIC_KEYS, () => HttpResponse.json({}, { status: 500 })),
    );
    const { result } = renderHookWithRouter(usePublicKeyList, {
      initialEntries: ["/?page=2"],
    });

    await waitFor(() => expect(result.current.error).toMatchObject({ status: 500 }));
    expect(result.current.params.page).toBe(2);
    expect(result.current.rows).toEqual([]);
  });
});
