import { useCallback, useRef } from "react";
import { useSearchParams } from "react-router-dom";
import {
  parseListParams,
  serializeListParams,
  type ListParamDefaults,
  type ListParamConstraints,
} from "./paginatedListParams";

/**
 * Describes a sortable field: which key it maps to and what order to use when
 * it is first selected (i.e. when the user switches from a different field).
 */
export interface SortFieldDef {
  /** The value that will appear in the `sortField` URL param. */
  field: string;
  /** The order to apply the first time this field is selected. */
  initialOrder: "asc" | "desc";
}

export interface UsePaginatedListStateConfig<T extends Record<string, unknown>> {
  defaults: ListParamDefaults<T>;
  constraints?: ListParamConstraints<T>;
  /**
   * Sort field definitions. When provided the hook exposes a `handleSort`
   * function. Each entry names a field and its `initialOrder` (the order used
   * when the user switches to that field for the first time).
   */
  sortFields?: SortFieldDef[];
  /**
   * Optional namespace prefix.  When set every URL key is written and read as
   * `<prefix>.<key>` so that two co-mounted instances with different prefixes
   * never clobber each other's state.
   */
  prefix?: string;
}

export interface UsePaginatedListStateResult<T extends Record<string, unknown>> {
  /** Parsed, typed params derived from the current URL. */
  params: T;
  /** The current URL search string (without leading "?"). Useful for assertions. */
  searchString: string;
  /** Navigate to a specific page. */
  setPage: (page: number) => void;
  /** Update the search string and reset page to 1. */
  setSearch: (search: string) => void;
  /**
   * Update a scalar (non-array) filter param and reset page to 1.
   * Invalid values (those rejected by the allowlist in `constraints`) will be
   * written to the URL and the parser will fall back to the default.
   */
  setFilter: <K extends keyof T>(key: K, value: T[K]) => void;
  /**
   * Update an array filter param and reset page to 1.
   * Values are serialized as repeated URL keys so that `getAll()` round-trips
   * correctly.
   */
  setArrayFilter: <K extends keyof T>(key: K, values: T[K]) => void;
  /**
   * Functionally update an array filter param by transforming the *committed*
   * URL value (not the render-closure snapshot). Use this instead of
   * `setArrayFilter` whenever the new array is derived from the current one —
   * it avoids stale-closure bugs when the array can change concurrently (e.g.
   * onTagRenamed / onTagDeleted in ManageTagsDrawer).
   */
  mapArrayFilter: <K extends keyof T>(
    key: K,
    fn: (current: T[K]) => T[K],
  ) => void;
  /**
   * Toggle the sort order when the same field is clicked, or switch to a
   * different field using its `initialOrder`. Resets page to 1.
   * Only available when `sortFields` was passed to the hook config.
   */
  handleSort: (field: string) => void;
  /** Reset all dimensions to their defaults. */
  reset: () => void;
}

// ── Internal helpers ──────────────────────────────────────────────────────────

/** Return `<prefix>.<key>` when a prefix is set, or the bare key otherwise. */
function prefixKey(key: string, prefix: string | undefined): string {
  return prefix ? `${prefix}.${key}` : key;
}

/**
 * Return an un-prefixed view of `full` that the pure parse/serialize helpers
 * can consume. For each managed key `k`, it copies every value stored under
 * `<prefix>.k` into the bare key `k`.  When there is no prefix the original
 * params are returned as-is.
 */
function stripPrefix<T extends Record<string, unknown>>(
  full: URLSearchParams,
  defaults: ListParamDefaults<T>,
  prefix: string | undefined,
): URLSearchParams {
  if (!prefix) return full;

  const stripped = new URLSearchParams();
  for (const key of Object.keys(defaults)) {
    for (const v of full.getAll(prefixKey(key, prefix))) {
      stripped.append(key, v);
    }
  }
  return stripped;
}

// ── Hook ──────────────────────────────────────────────────────────────────────

/**
 * Stable empty-constraints sentinel.  Using a module-level constant instead of
 * an inline `{}` default means callers that don't supply `constraints` all
 * share the same object reference, keeping the `update` useCallback stable
 * across renders.
 */
const EMPTY_CONSTRAINTS: ListParamConstraints<Record<string, unknown>> = {};

/**
 * Config-driven hook that syncs a paginated list's URL state (page, search,
 * sort, and any extra dimensions) with `useSearchParams`.
 *
 * Rules:
 * - Default-valued params are omitted from the URL (replace-history mode).
 * - Unrelated params already present in the URL are always preserved.
 * - Every non-page setter resets page to 1.
 * - When a `prefix` is given, every managed key is namespaced as
 *   `<prefix>.<key>` in the URL so multiple instances coexist cleanly.
 */
export function usePaginatedListState<T extends Record<string, unknown>>({
  defaults,
  constraints = EMPTY_CONSTRAINTS as ListParamConstraints<T>,
  sortFields,
  prefix,
}: UsePaginatedListStateConfig<T>): UsePaginatedListStateResult<T> {
  const [searchParams, setSearchParams] = useSearchParams();

  // Parse the current URL into typed params (handles prefix translation).
  // Thread the previous result through `prev` so array dimensions reuse their
  // prior reference when content-equal, keeping memoized deps stable. The ref
  // only caches a pure derived value, so reading/writing it during render is
  // safe — disable react-hooks/refs for these two intentional accesses.
  const paramsRef = useRef<T | undefined>(undefined);
  const stripped = stripPrefix(searchParams, defaults, prefix);
  // eslint-disable-next-line react-hooks/refs -- caches prior parse result for referential stability
  const params = parseListParams<T>(stripped, defaults, constraints, paramsRef.current);
  // eslint-disable-next-line react-hooks/refs -- stash this render's result for the next render
  paramsRef.current = params;

  /**
   * Merge the given partial update into the URL.
   *
   * - Unrelated params survive.
   * - Keys managed by THIS instance are written with the prefix.
   * - Default-valued keys are omitted.
   */
  const update = useCallback(
    (patch: Partial<T>) => {
      setSearchParams(
        (prev) => {
          // Start from a copy of the current URL so unrelated params survive.
          const next = new URLSearchParams(prev);

          // Apply the patch onto the current parsed (un-prefixed) state.
          const currentStripped = stripPrefix(prev, defaults, prefix);
          const current = parseListParams<T>(currentStripped, defaults, constraints);
          const merged = { ...current, ...patch };

          // Serialize the merged state (default-valued keys are omitted).
          const serialized = serializeListParams<T>(merged, defaults);

          // Remove every key managed by this instance (prefixed) from `next`,
          // then re-add only the non-default ones from `serialized`.
          for (const key of Object.keys(defaults)) {
            const urlKey = prefixKey(key, prefix);
            next.delete(urlKey);
            const newValues = serialized.getAll(key);
            for (const v of newValues) {
              next.append(urlKey, v);
            }
          }

          return next;
        },
        { replace: true },
      );
    },
    [setSearchParams, defaults, constraints, prefix],
  );

  const setPage = useCallback(
    (page: number) => {
      update({ page } as unknown as Partial<T>);
    },
    [update],
  );

  const setSearch = useCallback(
    (search: string) => {
      update({ search, page: defaults.page } as unknown as Partial<T>);
    },
    [update, defaults.page],
  );

  const setFilter = useCallback(
    <K extends keyof T>(key: K, value: T[K]) => {
      update({ [key]: value, page: defaults.page } as unknown as Partial<T>);
    },
    [update, defaults.page],
  );

  const setArrayFilter = useCallback(
    <K extends keyof T>(key: K, values: T[K]) => {
      update({ [key]: values, page: defaults.page } as unknown as Partial<T>);
    },
    [update, defaults.page],
  );

  const mapArrayFilter = useCallback(
    <K extends keyof T>(key: K, fn: (current: T[K]) => T[K]) => {
      // Compute the new array inside the setSearchParams callback so we always
      // read the *committed* URL state, never a stale render-closure snapshot.
      setSearchParams(
        (prev) => {
          const currentStripped = stripPrefix(prev, defaults, prefix);
          const current = parseListParams<T>(currentStripped, defaults, constraints);
          const next = fn(current[key]);
          const merged = { ...current, [key]: next, page: defaults.page };
          const serialized = serializeListParams<T>(merged, defaults);

          const nextParams = new URLSearchParams(prev);
          for (const k of Object.keys(defaults)) {
            const urlKey = prefixKey(k, prefix);
            nextParams.delete(urlKey);
            const newValues = serialized.getAll(k);
            for (const v of newValues) {
              nextParams.append(urlKey, v);
            }
          }
          return nextParams;
        },
        { replace: true },
      );
    },
    [setSearchParams, defaults, constraints, prefix],
  );

  const handleSort = useCallback(
    (field: string) => {
      // Derive currentField / currentOrder inside the setSearchParams callback
      // so we always read the *committed* URL state, never a stale render closure.
      setSearchParams(
        (prev) => {
          const currentStripped = stripPrefix(prev, defaults, prefix);
          const current = parseListParams<T>(currentStripped, defaults, constraints);
          // Access sortField / sortOrder through a type-erased view — they are
          // present in T only when the caller includes them in their params type.
          const p = current as Record<string, unknown>;
          const currentField = p["sortField"] as string | undefined;
          const currentOrder = p["sortOrder"] as "asc" | "desc" | undefined;

          const nextOrder: "asc" | "desc" =
            field === currentField
              ? // Same field — toggle.
                currentOrder === "asc"
                ? "desc"
                : "asc"
              : // New field — use its declared initialOrder, falling back to "asc".
                (sortFields?.find((f) => f.field === field)?.initialOrder ?? "asc");

          const patch = {
            sortField: field,
            sortOrder: nextOrder,
            page: defaults.page,
          } as unknown as Partial<T>;

          const merged = { ...current, ...patch };
          const serialized = serializeListParams<T>(merged, defaults);

          const next = new URLSearchParams(prev);
          for (const key of Object.keys(defaults)) {
            const urlKey = prefixKey(key, prefix);
            next.delete(urlKey);
            const newValues = serialized.getAll(key);
            for (const v of newValues) {
              next.append(urlKey, v);
            }
          }
          return next;
        },
        { replace: true },
      );
    },
    [setSearchParams, defaults, constraints, prefix, sortFields],
  );

  const reset = useCallback(() => {
    update({ ...defaults });
  }, [update, defaults]);

  return {
    params,
    searchString: searchParams.toString(),
    setPage,
    setSearch,
    setFilter,
    setArrayFilter,
    mapArrayFilter,
    handleSort,
    reset,
  };
}
