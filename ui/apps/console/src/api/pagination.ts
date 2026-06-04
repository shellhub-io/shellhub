export interface PaginatedResult<T> {
  data: T[];
  totalCount: number;
}

type SdkListFn<T, O> = (
  options: O & { throwOnError: true },
) => Promise<{ data: T[]; response: Response }>;

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
