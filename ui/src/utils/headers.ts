export const parseTotalCount = (headers: Record<string, unknown>): number =>
  parseInt(headers["x-total-count"] as string, 10) || 0;
