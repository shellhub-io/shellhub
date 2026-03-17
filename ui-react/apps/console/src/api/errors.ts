/**
 * Shape attached to errors by the fetch error interceptor in fetchInterceptors.ts.
 * The interceptor monkey-patches `.status` and `.headers` onto the parsed
 * response body before it is thrown by the SDK with `throwOnError: true`.
 */
export interface SdkHttpError {
  status: number;
  headers: Headers;
}

/**
 * Type guard analogous to `axios.isAxiosError()` — narrows an unknown
 * caught value to an object carrying the HTTP status and response headers
 * that the fetch error interceptor attaches.
 */
export function isSdkError(err: unknown): err is SdkHttpError {
  return (
    typeof err === "object"
    && err !== null
    && "status" in err
    && typeof err.status === "number"
  );
}
