export interface PaginatedResult<T> {
  data: T[];
  totalCount: number;
}

export function paginatedQueryFn<T>(
  sdkFn: (options: never) => Promise<{ data?: T[]; error?: unknown; response: Response }>,
  options: unknown,
): () => Promise<PaginatedResult<T>> {
  return async () => {
    const result = await (sdkFn as (options: unknown) => Promise<{ data?: T[]; error?: unknown; response: Response }>)(options);
    if (result.error) {
      throw result.error instanceof Error ? result.error : new Error(JSON.stringify(result.error));
    }
    const totalCount = parseInt(result.response.headers.get("X-Total-Count") ?? "0", 10);
    return { data: result.data ?? [], totalCount };
  };
}
