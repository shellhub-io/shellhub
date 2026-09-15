/**
 * Reads the non-enumerable `totalCount` that `customInstance` attaches to array responses carrying
 * an `X-Total-Count` header. Returns 0 when the data is undefined or the property is absent.
 */
export function totalCount(data: unknown[] | undefined): number {
  if (!data) return 0;
  return (data as unknown as { totalCount?: number }).totalCount ?? 0;
}
