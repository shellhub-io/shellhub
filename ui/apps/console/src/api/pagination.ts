import { fetchWithHeaders } from "./customInstance";

/**
 * A page of results together with the total the filter matched, which is what the pager needs
 * and the page itself cannot say.
 */
export interface PaginatedResult<T> {
  data: T[];
  totalCount: number;
}

/**
 * Wraps a generated list call as a query function that also reads the total from X-Total-Count.
 * The count lives in a header rather than the body, so a plain SDK call cannot page; every
 * paginated hook goes through here instead of parsing the header again.
 */
export function paginatedQueryFn<T>(
  url: string,
): (ctx: { signal: AbortSignal }) => Promise<PaginatedResult<T>> {
  return async ({ signal }) => {
    const { data, headers } = await fetchWithHeaders<T[]>(url, { method: "GET", signal });
    const totalCount = parseInt(headers.get("X-Total-Count") ?? "0", 10);
    return { data, totalCount };
  };
}
