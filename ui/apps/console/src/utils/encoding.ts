import { Buffer } from "buffer";

/**
 * Serializes a value as unpadded base64url, for a filter carried in a query string or a header.
 *
 * btoa(JSON.stringify(value)) throws InvalidCharacterError on any character above U+00FF — a tag
 * name, hostname or username in a non-Latin script — so the request would never leave the browser.
 * The value is UTF-8 encoded first, then the standard base64 is rewritten as base64url so it
 * survives a URL without percent-encoding.
 *
 * The rewrite is done by hand because feross/buffer 5.7.1, which the lockfile pins, has no
 * toString("base64url").
 */
export function toBase64Json(value: unknown): string {
  return Buffer.from(JSON.stringify(value), "utf-8")
    .toString("base64")
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");
}

/**
 * One property condition of a list filter, matching the backend's `query.FilterProperty`. The route
 * answers 400 (`ErrFilterPropertyInvalid`) when `query.ValidateFilters` rejects the condition: a
 * name/operator pair its filter fields do not allow, or a value of the wrong type, size or set for
 * its field and operator.
 */
export function propertyFilter(
  name: string,
  operator: "contains" | "eq" | "ne" | "bool" | "gt" | "lt",
  value: unknown,
) {
  return { type: "property", params: { name, operator, value } } as const;
}

/**
 * An operator node of a list filter, matching the backend's `query.FilterOperator`. It joins every
 * condition after it, until the next operator; conditions before the first operator join with or.
 */
export function operatorFilter(name: "and" | "or") {
  return { type: "operator", params: { name } } as const;
}

/**
 * A group node of a list filter: the backend evaluates the nodes it holds as one bracketed
 * condition. A group cannot hold another group, must hold at least one property, and counts with
 * every node inside it toward the backend's cap of 8 nodes, or the route answers 400.
 */
export function groupFilter(
  ...nodes: (
    ReturnType<typeof propertyFilter> | ReturnType<typeof operatorFilter>
  )[]
) {
  return { type: "group", params: nodes } as const;
}

/**
 * The conditions with an or operator between each pair. Wrap them in {@link groupFilter} to join them
 * as one condition to the rest of the filter, since and binds tighter than or.
 */
export function orJoined(conditions: ReturnType<typeof propertyFilter>[]) {
  return conditions.flatMap((condition, i) =>
    i > 0 ? [operatorFilter("or"), condition] : [condition],
  );
}
