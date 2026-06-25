/**
 * Pure (no React, no router) helpers for reading and writing paginated list
 * parameters from/to a URLSearchParams instance.
 */

/**
 * A record that maps each key of T to its default value.
 * Used both to drive parsing fall-backs and to suppress default values during
 * serialization.
 */
export type ListParamDefaults<T extends Record<string, unknown>> = {
  [K in keyof T]: T[K];
};

/**
 * Validation constraints for individual dimensions.
 *
 * - For numeric scalars (page, perPage): an optional readonly array of allowed
 *   values. When absent the value is validated only as a positive integer.
 * - For string scalars: an optional readonly array of allowed values.
 * - For array dimensions: an optional readonly array; values not in the list
 *   are filtered out.
 * - Omit a key to skip allowlist validation for that dimension.
 */
export type ListParamConstraints<T extends Record<string, unknown>> = {
  [K in keyof T]?: T[K] extends unknown[]
    ? readonly T[K][number][]
    : readonly T[K][];
};

// ── Internal helpers ──────────────────────────────────────────────────────────

function parsePositiveInt(raw: string | null): number | null {
  if (raw === null) return null;
  const n = Number(raw);
  if (!Number.isInteger(n) || n <= 0) return null;
  return n;
}

// ── parseListParams ───────────────────────────────────────────────────────────

/**
 * Parse a URLSearchParams instance into a typed params object.
 *
 * Rules per dimension:
 * - `page`: positive integer; falls back to default on missing/invalid.
 * - `perPage`: positive integer; validated against `constraints.perPage` when
 *   provided; falls back to default on missing/invalid/disallowed.
 * - string scalar: raw value; validated against `constraints[key]` when
 *   provided; falls back to default when the value is not in the allowlist.
 * - array: all values for the key collected via `getAll()`; filtered against
 *   `constraints[key]` when provided.
 *
 * `prev` is an optional previous result. When provided, array dimensions that
 * are content-equal to the previous value reuse the same array reference so
 * that memoized React dependencies (useCallback, useMemo) remain stable across
 * renders that don't actually change the array's contents.
 */
export function parseListParams<T extends Record<string, unknown>>(
  searchParams: URLSearchParams,
  defaults: ListParamDefaults<T>,
  constraints: ListParamConstraints<T>,
  prev?: T,
): T {
  const result: Record<string, unknown> = {};

  for (const key of Object.keys(defaults) as (keyof T & string)[]) {
    const defaultValue = defaults[key];

    if (Array.isArray(defaultValue)) {
      // Array dimension
      const all = searchParams.getAll(key);
      const allowlist = constraints[key] as readonly string[] | undefined;
      // Deduplicate (preserving first-seen order) so e.g. ?tags=a&tags=a does
      // not round-trip a dirty URL or double-apply the filter.
      const values = [
        ...new Set(allowlist ? all.filter((v) => allowlist.includes(v)) : all),
      ];
      // Return the previous array reference when contents are identical so
      // that React memoization (useCallback/useMemo) deps stay stable.
      // Use set-equality (order-insensitive) so that e.g. ?tags=b&tags=a and
      // ?tags=a&tags=b produce the same reference — tag order is irrelevant for
      // filtering and reordering should not bust memoization.
      const prevArr = prev ? (prev[key] as unknown[]) : undefined;
      if (prevArr && prevArr.length === values.length) {
        const prevSet = new Set(prevArr);
        if (values.every((v) => prevSet.has(v))) {
          result[key] = prevArr;
          continue;
        }
      }
      result[key] = values;
      continue;
    }

    if (key === "page") {
      const raw = searchParams.get(key);
      const parsed = parsePositiveInt(raw);
      result[key] = parsed !== null ? parsed : defaultValue;
      continue;
    }

    if (typeof defaultValue === "number") {
      // Numeric scalar (e.g. perPage)
      const raw = searchParams.get(key);
      const parsed = parsePositiveInt(raw);
      if (parsed === null) {
        result[key] = defaultValue;
      } else {
        const allowlist = constraints[key] as readonly number[] | undefined;
        if (allowlist && !allowlist.includes(parsed)) {
          result[key] = defaultValue;
        } else {
          result[key] = parsed;
        }
      }
      continue;
    }

    // String scalar
    const raw = searchParams.get(key);
    if (raw === null) {
      result[key] = defaultValue;
      continue;
    }
    const allowlist = constraints[key] as readonly string[] | undefined;
    if (allowlist && !allowlist.includes(raw)) {
      result[key] = defaultValue;
    } else {
      result[key] = raw;
    }
  }

  return result as T;
}

// ── serializeListParams ───────────────────────────────────────────────────────

/**
 * Serialize a typed params object into a URLSearchParams instance.
 *
 * Params equal to their default are omitted so URLs stay clean.
 * Array dimensions are serialized as repeated keys (one entry per element).
 */
export function serializeListParams<T extends Record<string, unknown>>(
  value: T,
  defaults: ListParamDefaults<T>,
): URLSearchParams {
  const sp = new URLSearchParams();

  for (const key of Object.keys(defaults) as (keyof T & string)[]) {
    const current = value[key];
    const defaultValue = defaults[key];

    if (Array.isArray(current)) {
      const currentArr = current as unknown[];
      const defaultArr = defaultValue as unknown[];
      // Compare by JSON to handle same-order equality
      const isDefault =
        currentArr.length === defaultArr.length &&
        currentArr.every((v, i) => v === defaultArr[i]);
      if (!isDefault) {
        for (const item of currentArr) {
          sp.append(key, String(item));
        }
      }
      continue;
    }

    if (current !== defaultValue) {
      sp.set(key, String(current));
    }
  }

  return sp;
}
