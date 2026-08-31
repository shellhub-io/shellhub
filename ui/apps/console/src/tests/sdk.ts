import type { SdkHttpError } from "@/api/errors";

/**
 * The shape the generated SDK resolves to, so a mocked call is typed as the real one.
 */
export type SdkResponse<T = unknown> = {
  data: T;
  error: undefined;
  request: Request;
  response: Response;
};

/**
 * Builds a successful SDK response around data.
 */
export function mockSdkResponse<T>(
  data: T,
  headers?: HeadersInit,
): SdkResponse<T> {
  return {
    data,
    error: undefined,
    request: new Request("http://localhost"),
    response: new Response(null, headers ? { headers } : undefined),
  };
}

/**
 * Builds a paginated SDK response, with the total in X-Total-Count where the real API puts it —
 * a paginated hook reads the header, not the body.
 */
export function paginatedResponse<T>(data: T[], total = data.length) {
  return mockSdkResponse(data, { "X-Total-Count": String(total) });
}

/**
 * Builds an SDK error with a status, for testing the paths that branch on one.
 */
export function makeSdkError(
  status: number,
  headers?: HeadersInit,
): Error & SdkHttpError {
  return Object.assign(new Error("Request failed"), {
    status,
    headers: new Headers(headers),
  });
}
