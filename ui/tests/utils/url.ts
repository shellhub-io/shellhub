/**
 * Builds a URL with query parameters from an object.
 *
 * @param baseUrl - The base URL (e.g., "http://localhost:3000/api/devices")
 * @param params - Optional object with query parameters
 * @returns The complete URL with query string
 *
 * @example
 * buildUrl("http://localhost:3000/api/devices", { page: "1", per_page: "10" })
 * // Returns: "http://localhost:3000/api/devices?page=1&per_page=10"
 */
export const buildUrl = (
  baseUrl: string,
  params?: Record<string, string | number>,
): string => {
  if (!params) return baseUrl;

  const queryParams = new URLSearchParams(params as Record<string, string>).toString();
  return `${baseUrl}?${queryParams}`;
};
