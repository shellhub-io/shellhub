/**
 * A page of results together with the total the filter matched, which is what the pager needs
 * and the page itself cannot say.
 */
export interface PaginatedResult<T> {
  data: T[];
  totalCount: number;
}

type SdkListFn<T, O> = (
  options: O & { throwOnError: true },
) => Promise<{ data: T[]; response: Response }>;

/**
 * Wraps a generated list call as a query function that also reads the total from X-Total-Count.
 * The count lives in a header rather than the body, so a plain SDK call cannot page; every
 * paginated hook goes through here instead of parsing the header again.
 */
export function paginatedQueryFn<T, O>(
  sdkFn: SdkListFn<T, O>,
  options: O,
): () => Promise<PaginatedResult<T>> {
  return async () => {
    const { data, response } = await sdkFn({ ...options, throwOnError: true });
    const totalCount = parseInt(response.headers.get("X-Total-Count") ?? "0", 10);
    return { data, totalCount };
  };
}
